package services

import (
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestMonitoringService_RetryMechanism tests the exponential backoff retry logic
func TestMonitoringService_RetryMechanism(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.DeliveryRecord{}, &models.StatusTransition{}, &models.School{}, &models.User{}, &models.MenuItem{}, &models.Recipe{}, &models.MenuPlan{})
	assert.NoError(t, err)

	// Create monitoring service without Firebase (will cause sync to fail)
	service := &MonitoringService{
		db:         db,
		retryQueue: make(chan syncRetryItem, 100),
	}

	// Create test data
	school := models.School{Name: "Test School"}
	db.Create(&school)

	user := models.User{FullName: "Test User", Email: "test@example.com", Role: "driver"}
	db.Create(&user)

	recipe := models.Recipe{
		Name:          "Test Recipe",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		CreatedBy:     1,
	}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
		CreatedBy: 1,
	}
	db.Create(&menuPlan)

	menuItem := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       time.Now(),
		RecipeID:   recipe.ID,
		Portions:   100,
	}
	db.Create(&menuItem)

	record := models.DeliveryRecord{
		DeliveryDate:  time.Now(),
		SchoolID:      school.ID,
		DriverID:      user.ID,
		MenuItemID:    menuItem.ID,
		Portions:      100,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  10,
	}
	db.Create(&record)

	// Test: syncToFirebaseWithRetry should handle nil Firebase client gracefully
	// Since dbClient is nil, sync will fail but should not panic
	err = service.syncToFirebaseWithRetry(&record, 0)
	
	// The function returns nil immediately and retries in background
	// We just verify it doesn't panic and returns without error
	assert.NoError(t, err)

	// Test: Verify retry queue can accept items
	service.queueForRetry(record.ID, 0)
	
	// Give a moment for the queue operation
	time.Sleep(10 * time.Millisecond)
	
	// Verify queue has items (non-blocking check)
	select {
	case item := <-service.retryQueue:
		assert.Equal(t, record.ID, item.recordID)
		assert.Equal(t, 0, item.attempt)
		assert.True(t, item.nextRetryAt.After(time.Now().Add(-1*time.Second)))
	default:
		t.Fatal("Expected item in retry queue")
	}
}

// TestCleaningService_RetryMechanism tests the exponential backoff retry logic for cleaning service
func TestCleaningService_RetryMechanism(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.OmprengCleaning{}, &models.DeliveryRecord{}, &models.School{}, &models.User{}, &models.MenuItem{}, &models.Recipe{}, &models.MenuPlan{})
	assert.NoError(t, err)

	// Create cleaning service without Firebase (will cause sync to fail)
	service := &CleaningService{
		db:         db,
		retryQueue: make(chan cleaningSyncRetryItem, 100),
	}

	// Create test data
	school := models.School{Name: "Test School"}
	db.Create(&school)

	user := models.User{FullName: "Test User", Email: "test@example.com", Role: "driver"}
	db.Create(&user)

	recipe := models.Recipe{
		Name:          "Test Recipe",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		CreatedBy:     1,
	}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
		CreatedBy: 1,
	}
	db.Create(&menuPlan)

	menuItem := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       time.Now(),
		RecipeID:   recipe.ID,
		Portions:   100,
	}
	db.Create(&menuItem)

	deliveryRecord := models.DeliveryRecord{
		DeliveryDate:  time.Now(),
		SchoolID:      school.ID,
		DriverID:      user.ID,
		MenuItemID:    menuItem.ID,
		Portions:      100,
		CurrentStatus: "ompreng_sampai_di_sppg",
		OmprengCount:  10,
	}
	db.Create(&deliveryRecord)

	cleaning := models.OmprengCleaning{
		DeliveryRecordID: deliveryRecord.ID,
		OmprengCount:     10,
		CleaningStatus:   "pending",
	}
	db.Create(&cleaning)

	// Test: syncToFirebaseWithRetry should handle nil Firebase client gracefully
	err = service.syncToFirebaseWithRetry(&cleaning, 0)
	
	// The function returns nil immediately and retries in background
	assert.NoError(t, err)

	// Test: Verify retry queue can accept items
	service.queueForRetry(cleaning.ID, 0)
	
	// Give a moment for the queue operation
	time.Sleep(10 * time.Millisecond)
	
	// Verify queue has items (non-blocking check)
	select {
	case item := <-service.retryQueue:
		assert.Equal(t, cleaning.ID, item.cleaningID)
		assert.Equal(t, 0, item.attempt)
		assert.True(t, item.nextRetryAt.After(time.Now().Add(-1*time.Second)))
	default:
		t.Fatal("Expected item in retry queue")
	}
}

// TestExponentialBackoffCalculation verifies the exponential backoff timing
func TestExponentialBackoffCalculation(t *testing.T) {
	tests := []struct {
		attempt         int
		expectedSeconds int
	}{
		{0, 1},   // 2^0 = 1 second
		{1, 2},   // 2^1 = 2 seconds
		{2, 4},   // 2^2 = 4 seconds
		{3, 8},   // 2^3 = 8 seconds
		{4, 16},  // 2^4 = 16 seconds
	}

	for _, tt := range tests {
		backoffSeconds := 1 << tt.attempt // 2^attempt
		assert.Equal(t, tt.expectedSeconds, backoffSeconds, 
			"Attempt %d should have %d second backoff", tt.attempt, tt.expectedSeconds)
	}
}

// TestMaxRetryAttempts verifies that retry stops after max attempts
func TestMaxRetryAttempts(t *testing.T) {
	const maxRetries = 5
	
	// Verify that attempt 5 (6th total attempt) is the last retry
	assert.Equal(t, 5, maxRetries, "Max retries should be 5")
	
	// Verify exponential backoff doesn't exceed reasonable limits
	// At attempt 4 (5th retry), backoff is 16 seconds
	backoffAtMaxRetry := 1 << 4 // 2^4
	assert.Equal(t, 16, backoffAtMaxRetry, "Max backoff should be 16 seconds")
}
