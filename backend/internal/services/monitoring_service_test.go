package services

import (
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMonitoringTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(
		&models.DeliveryRecord{},
		&models.School{},
		&models.User{},
		&models.Recipe{},
		&models.MenuPlan{},
		&models.MenuItem{},
		&models.StatusTransition{},
		&models.OmprengCleaning{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// Helper function to create a monitoring service for tests (without Firebase)
func newTestMonitoringService(db *gorm.DB) *MonitoringService {
	return &MonitoringService{
		db:          db,
		firebaseApp: nil,
		dbClient:    nil,
		retryQueue:  make(chan syncRetryItem, 100),
	}
}

func TestMonitoringService_GetDeliveryRecordDetail(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     driver.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: driver.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Test: Get delivery record detail
	result, err := service.GetDeliveryRecordDetail(deliveryRecord.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, deliveryRecord.ID, result.ID)
	assert.Equal(t, deliveryDate.Unix(), result.DeliveryDate.Unix())
	assert.Equal(t, 150, result.Portions)
	assert.Equal(t, "sedang_dimasak", result.CurrentStatus)
	assert.Equal(t, 15, result.OmprengCount)

	// Verify School association is preloaded
	assert.NotNil(t, result.School)
	assert.Equal(t, school.ID, result.School.ID)
	assert.Equal(t, "SD Negeri 1", result.School.Name)
	assert.Equal(t, "Jl. Test No. 1", result.School.Address)
	assert.Equal(t, "John Doe", result.School.ContactPerson)
	assert.Equal(t, "081234567890", result.School.PhoneNumber)

	// Verify Driver association is preloaded
	assert.NotNil(t, result.Driver)
	assert.Equal(t, driver.ID, result.Driver.ID)
	assert.Equal(t, "Driver Test", result.Driver.FullName)
	assert.Equal(t, "driver", result.Driver.Role)

	// Verify MenuItem association is preloaded
	assert.NotNil(t, result.MenuItem)
	assert.Equal(t, menuItem.ID, result.MenuItem.ID)
	assert.Equal(t, 150, result.MenuItem.Portions)
}

func TestMonitoringService_GetDeliveryRecordDetail_NotFound(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Test: Get non-existent delivery record
	result, err := service.GetDeliveryRecordDetail(999)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestMonitoringService_GetDeliveryRecordDetail_MultipleRecords(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 2",
		Address:       "Jl. Test No. 2",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "Jane Doe",
		PhoneNumber:   "081234567891",
		StudentCount:  200,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "0987654321",
		Email:        "driver2@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Two",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Kuning",
		Category:      "main",
		TotalCalories: 450,
		TotalProtein:  18,
		TotalCarbs:    55,
		TotalFat:      12,
		IsActive:      true,
		CreatedBy:     driver.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: driver.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   100,
	}
	db.Create(menuItem)

	// Create multiple delivery records
	record1 := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      100,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  10,
	}
	db.Create(record1)

	record2 := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      200,
		CurrentStatus: "diperjalanan",
		OmprengCount:  20,
	}
	db.Create(record2)

	// Test: Get specific delivery record (record2)
	result, err := service.GetDeliveryRecordDetail(record2.ID)

	// Assertions - should return only record2, not record1
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, record2.ID, result.ID)
	assert.Equal(t, 200, result.Portions)
	assert.Equal(t, "diperjalanan", result.CurrentStatus)
	assert.Equal(t, 20, result.OmprengCount)
	assert.NotEqual(t, record1.ID, result.ID)
}

func TestMonitoringService_UpdateDeliveryStatus_Success(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test chef
	chef := &models.User{
		NIK:          "0987654321",
		Email:        "chef@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Chef Test",
		Role:         "chef",
		IsActive:     true,
	}
	db.Create(chef)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     chef.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: chef.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record with initial status
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Test: Update status from "sedang_dimasak" to "selesai_dimasak"
	err := service.UpdateDeliveryStatus(deliveryRecord.ID, "selesai_dimasak", chef.ID, "Cooking completed")

	// Assertions
	assert.NoError(t, err)

	// Verify delivery record was updated
	var updatedRecord models.DeliveryRecord
	db.First(&updatedRecord, deliveryRecord.ID)
	assert.Equal(t, "selesai_dimasak", updatedRecord.CurrentStatus)

	// Verify status transition was created
	var transition models.StatusTransition
	err = db.Where("delivery_record_id = ?", deliveryRecord.ID).First(&transition).Error
	assert.NoError(t, err)
	assert.Equal(t, deliveryRecord.ID, transition.DeliveryRecordID)
	assert.Equal(t, "sedang_dimasak", transition.FromStatus)
	assert.Equal(t, "selesai_dimasak", transition.ToStatus)
	assert.Equal(t, chef.ID, transition.TransitionedBy)
	assert.Equal(t, "Cooking completed", transition.Notes)
	assert.NotZero(t, transition.TransitionedAt)
}

func TestMonitoringService_UpdateDeliveryStatus_InvalidTransition(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test chef
	chef := &models.User{
		NIK:          "0987654321",
		Email:        "chef@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Chef Test",
		Role:         "chef",
		IsActive:     true,
	}
	db.Create(chef)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     chef.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: chef.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record with initial status
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Test: Try to update status from "sedang_dimasak" to "siap_dipacking" (invalid - skips "selesai_dimasak")
	err := service.UpdateDeliveryStatus(deliveryRecord.ID, "siap_dipacking", chef.ID, "")

	// Assertions
	assert.Error(t, err)
	assert.IsType(t, &InvalidTransitionError{}, err)

	// Verify delivery record was NOT updated
	var unchangedRecord models.DeliveryRecord
	db.First(&unchangedRecord, deliveryRecord.ID)
	assert.Equal(t, "sedang_dimasak", unchangedRecord.CurrentStatus)

	// Verify no status transition was created
	var count int64
	db.Model(&models.StatusTransition{}).Where("delivery_record_id = ?", deliveryRecord.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestMonitoringService_UpdateDeliveryStatus_RecordNotFound(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Test: Try to update non-existent delivery record
	err := service.UpdateDeliveryStatus(999, "selesai_dimasak", 1, "")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestMonitoringService_UpdateDeliveryStatus_MultipleTransitions(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test chef
	chef := &models.User{
		NIK:          "0987654321",
		Email:        "chef@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Chef Test",
		Role:         "chef",
		IsActive:     true,
	}
	db.Create(chef)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     chef.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: chef.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record with initial status
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Test: Perform multiple status transitions
	err := service.UpdateDeliveryStatus(deliveryRecord.ID, "selesai_dimasak", chef.ID, "Cooking completed")
	assert.NoError(t, err)

	err = service.UpdateDeliveryStatus(deliveryRecord.ID, "siap_dipacking", chef.ID, "Ready for packing")
	assert.NoError(t, err)

	err = service.UpdateDeliveryStatus(deliveryRecord.ID, "selesai_dipacking", chef.ID, "Packing completed")
	assert.NoError(t, err)

	// Verify final status
	var finalRecord models.DeliveryRecord
	db.First(&finalRecord, deliveryRecord.ID)
	assert.Equal(t, "selesai_dipacking", finalRecord.CurrentStatus)

	// Verify all transitions were created
	var transitions []models.StatusTransition
	db.Where("delivery_record_id = ?", deliveryRecord.ID).Order("transitioned_at ASC").Find(&transitions)
	assert.Equal(t, 3, len(transitions))

	// Verify transition sequence
	assert.Equal(t, "sedang_dimasak", transitions[0].FromStatus)
	assert.Equal(t, "selesai_dimasak", transitions[0].ToStatus)

	assert.Equal(t, "selesai_dimasak", transitions[1].FromStatus)
	assert.Equal(t, "siap_dipacking", transitions[1].ToStatus)

	assert.Equal(t, "siap_dipacking", transitions[2].FromStatus)
	assert.Equal(t, "selesai_dipacking", transitions[2].ToStatus)
}

func TestMonitoringService_GetActivityLog_Success(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test chef
	chef := &models.User{
		NIK:          "0987654321",
		Email:        "chef@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Chef Test",
		Role:         "chef",
		IsActive:     true,
	}
	db.Create(chef)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     chef.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: chef.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Create status transitions manually with specific timestamps
	transition1 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "",
		ToStatus:         "sedang_dimasak",
		TransitionedAt:   time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
		TransitionedBy:   chef.ID,
		Notes:            "Started cooking",
	}
	db.Create(transition1)

	transition2 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "sedang_dimasak",
		ToStatus:         "selesai_dimasak",
		TransitionedAt:   time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
		TransitionedBy:   chef.ID,
		Notes:            "Cooking completed",
	}
	db.Create(transition2)

	transition3 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "selesai_dimasak",
		ToStatus:         "siap_dipacking",
		TransitionedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		TransitionedBy:   chef.ID,
		Notes:            "Ready for packing",
	}
	db.Create(transition3)

	// Test: Get activity log
	activityLog, err := service.GetActivityLog(deliveryRecord.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, activityLog)
	assert.Equal(t, 3, len(activityLog))

	// Verify chronological order (ASC)
	assert.Equal(t, transition1.ID, activityLog[0].ID)
	assert.Equal(t, transition2.ID, activityLog[1].ID)
	assert.Equal(t, transition3.ID, activityLog[2].ID)

	// Verify first transition
	assert.Equal(t, "", activityLog[0].FromStatus)
	assert.Equal(t, "sedang_dimasak", activityLog[0].ToStatus)
	assert.Equal(t, "Started cooking", activityLog[0].Notes)
	assert.Equal(t, chef.ID, activityLog[0].TransitionedBy)
	assert.NotNil(t, activityLog[0].User)
	assert.Equal(t, "Chef Test", activityLog[0].User.FullName)
	assert.Equal(t, "chef", activityLog[0].User.Role)

	// Verify second transition
	assert.Equal(t, "sedang_dimasak", activityLog[1].FromStatus)
	assert.Equal(t, "selesai_dimasak", activityLog[1].ToStatus)
	assert.Equal(t, "Cooking completed", activityLog[1].Notes)
	assert.Equal(t, chef.ID, activityLog[1].TransitionedBy)
	assert.NotNil(t, activityLog[1].User)
	assert.Equal(t, "Chef Test", activityLog[1].User.FullName)

	// Verify third transition
	assert.Equal(t, "selesai_dimasak", activityLog[2].FromStatus)
	assert.Equal(t, "siap_dipacking", activityLog[2].ToStatus)
	assert.Equal(t, "Ready for packing", activityLog[2].Notes)

	// Verify elapsed time can be calculated
	elapsed1 := activityLog[1].TransitionedAt.Sub(activityLog[0].TransitionedAt)
	assert.Equal(t, 90*time.Minute, elapsed1) // 1.5 hours

	elapsed2 := activityLog[2].TransitionedAt.Sub(activityLog[1].TransitionedAt)
	assert.Equal(t, 30*time.Minute, elapsed2) // 30 minutes
}

func TestMonitoringService_GetActivityLog_EmptyLog(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     driver.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: driver.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record with no transitions
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Test: Get activity log for record with no transitions
	activityLog, err := service.GetActivityLog(deliveryRecord.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, activityLog)
	assert.Equal(t, 0, len(activityLog))
}

func TestMonitoringService_GetActivityLog_MultipleUsers(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test users
	chef := &models.User{
		NIK:          "1111111111",
		Email:        "chef@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Chef Test",
		Role:         "chef",
		IsActive:     true,
	}
	db.Create(chef)

	packing := &models.User{
		NIK:          "2222222222",
		Email:        "packing@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Packing Staff",
		Role:         "packing",
		IsActive:     true,
	}
	db.Create(packing)

	driver := &models.User{
		NIK:          "3333333333",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     chef.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: chef.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create test delivery record
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	}
	db.Create(deliveryRecord)

	// Create transitions by different users
	transition1 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "",
		ToStatus:         "sedang_dimasak",
		TransitionedAt:   time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
		TransitionedBy:   chef.ID,
		Notes:            "Started cooking",
	}
	db.Create(transition1)

	transition2 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "sedang_dimasak",
		ToStatus:         "selesai_dimasak",
		TransitionedAt:   time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
		TransitionedBy:   chef.ID,
		Notes:            "Cooking completed",
	}
	db.Create(transition2)

	transition3 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "selesai_dimasak",
		ToStatus:         "siap_dipacking",
		TransitionedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		TransitionedBy:   packing.ID,
		Notes:            "Ready for packing",
	}
	db.Create(transition3)

	transition4 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "siap_dipacking",
		ToStatus:         "selesai_dipacking",
		TransitionedAt:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		TransitionedBy:   packing.ID,
		Notes:            "Packing completed",
	}
	db.Create(transition4)

	transition5 := &models.StatusTransition{
		DeliveryRecordID: deliveryRecord.ID,
		FromStatus:       "selesai_dipacking",
		ToStatus:         "siap_dikirim",
		TransitionedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		TransitionedBy:   driver.ID,
		Notes:            "Ready for delivery",
	}
	db.Create(transition5)

	// Test: Get activity log
	activityLog, err := service.GetActivityLog(deliveryRecord.ID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, activityLog)
	assert.Equal(t, 5, len(activityLog))

	// Verify User associations are preloaded with correct users
	assert.Equal(t, "Chef Test", activityLog[0].User.FullName)
	assert.Equal(t, "chef", activityLog[0].User.Role)

	assert.Equal(t, "Chef Test", activityLog[1].User.FullName)
	assert.Equal(t, "chef", activityLog[1].User.Role)

	assert.Equal(t, "Packing Staff", activityLog[2].User.FullName)
	assert.Equal(t, "packing", activityLog[2].User.Role)

	assert.Equal(t, "Packing Staff", activityLog[3].User.FullName)
	assert.Equal(t, "packing", activityLog[3].User.Role)

	assert.Equal(t, "Driver Test", activityLog[4].User.FullName)
	assert.Equal(t, "driver", activityLog[4].User.Role)
}

func TestMonitoringService_GetDailySummary_Success(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test schools
	school1 := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school1)

	school2 := &models.School{
		Name:          "SD Negeri 2",
		Address:       "Jl. Test No. 2",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "Jane Doe",
		PhoneNumber:   "081234567891",
		StudentCount:  200,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school2)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test delivery date
	deliveryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     driver.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: driver.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       deliveryDate,
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create delivery records with various statuses
	// 2 records in cooking stage
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school1.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school2.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      200,
		CurrentStatus: "selesai_dimasak",
		OmprengCount:  20,
	})

	// 3 records completed
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school1.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      100,
		CurrentStatus: "sudah_diterima_pihak_sekolah",
		OmprengCount:  10,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school2.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      120,
		CurrentStatus: "sudah_diterima_pihak_sekolah",
		OmprengCount:  12,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school1.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      80,
		CurrentStatus: "sudah_diterima_pihak_sekolah",
		OmprengCount:  8,
	})

	// 2 records in cleaning
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school2.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      90,
		CurrentStatus: "ompreng_proses_pencucian",
		OmprengCount:  9,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school1.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      110,
		CurrentStatus: "ompreng_proses_pencucian",
		OmprengCount:  11,
	})

	// 1 record cleaned
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  deliveryDate,
		SchoolID:      school2.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      130,
		CurrentStatus: "ompreng_selesai_dicuci",
		OmprengCount:  13,
	})

	// Test: Get daily summary
	summary, err := service.GetDailySummary(deliveryDate)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Verify total deliveries count
	assert.Equal(t, 8, summary.TotalDeliveries)

	// Verify completed deliveries count
	assert.Equal(t, 3, summary.CompletedDeliveries)

	// Verify ompreng in cleaning count
	assert.Equal(t, 2, summary.OmprengInCleaning)

	// Verify ompreng cleaned count
	assert.Equal(t, 1, summary.OmprengCleaned)

	// Verify status counts map
	assert.NotNil(t, summary.StatusCounts)
	assert.Equal(t, 1, summary.StatusCounts["sedang_dimasak"])
	assert.Equal(t, 1, summary.StatusCounts["selesai_dimasak"])
	assert.Equal(t, 3, summary.StatusCounts["sudah_diterima_pihak_sekolah"])
	assert.Equal(t, 2, summary.StatusCounts["ompreng_proses_pencucian"])
	assert.Equal(t, 1, summary.StatusCounts["ompreng_selesai_dicuci"])
}

func TestMonitoringService_GetDailySummary_EmptyDate(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Test: Get summary for date with no deliveries
	emptyDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	summary, err := service.GetDailySummary(emptyDate)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 0, summary.TotalDeliveries)
	assert.Equal(t, 0, summary.CompletedDeliveries)
	assert.Equal(t, 0, summary.OmprengInCleaning)
	assert.Equal(t, 0, summary.OmprengCleaned)
	assert.NotNil(t, summary.StatusCounts)
	assert.Equal(t, 0, len(summary.StatusCounts))
}

func TestMonitoringService_GetDailySummary_DifferentDates(t *testing.T) {
	db := setupMonitoringTestDB(t)
	service := newTestMonitoringService(db)

	// Create test school
	school := &models.School{
		Name:          "SD Negeri 1",
		Address:       "Jl. Test No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "John Doe",
		PhoneNumber:   "081234567890",
		StudentCount:  150,
		Category:      "SD",
		IsActive:      true,
	}
	db.Create(school)

	// Create test driver
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	db.Create(driver)

	// Create test recipe
	recipe := &models.Recipe{
		Name:          "Nasi Goreng",
		Category:      "main",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		IsActive:      true,
		CreatedBy:     driver.ID,
	}
	db.Create(recipe)

	// Create test menu plan
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)
	menuPlan := &models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: driver.ID,
	}
	db.Create(menuPlan)

	// Create test menu item
	menuItem := &models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		RecipeID:   recipe.ID,
		Portions:   150,
	}
	db.Create(menuItem)

	// Create delivery records for different dates
	date1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	// 3 records for date1
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  date1,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      150,
		CurrentStatus: "sedang_dimasak",
		OmprengCount:  15,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  date1,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      100,
		CurrentStatus: "sudah_diterima_pihak_sekolah",
		OmprengCount:  10,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  date1,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      120,
		CurrentStatus: "ompreng_proses_pencucian",
		OmprengCount:  12,
	})

	// 2 records for date2
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  date2,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      200,
		CurrentStatus: "selesai_dimasak",
		OmprengCount:  20,
	})
	db.Create(&models.DeliveryRecord{
		DeliveryDate:  date2,
		SchoolID:      school.ID,
		DriverID:      driver.ID,
		MenuItemID:    menuItem.ID,
		Portions:      180,
		CurrentStatus: "ompreng_selesai_dicuci",
		OmprengCount:  18,
	})

	// Test: Get summary for date1
	summary1, err := service.GetDailySummary(date1)
	assert.NoError(t, err)
	assert.NotNil(t, summary1)
	assert.Equal(t, 3, summary1.TotalDeliveries)
	assert.Equal(t, 1, summary1.CompletedDeliveries)
	assert.Equal(t, 1, summary1.OmprengInCleaning)
	assert.Equal(t, 0, summary1.OmprengCleaned)

	// Test: Get summary for date2
	summary2, err := service.GetDailySummary(date2)
	assert.NoError(t, err)
	assert.NotNil(t, summary2)
	assert.Equal(t, 2, summary2.TotalDeliveries)
	assert.Equal(t, 0, summary2.CompletedDeliveries)
	assert.Equal(t, 0, summary2.OmprengInCleaning)
	assert.Equal(t, 1, summary2.OmprengCleaned)
}
