package services

import (
	"context"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupActivityTrackerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.School{},
		&models.DeliveryRecord{},
		&models.StatusTransition{},
	)
	require.NoError(t, err)

	return db
}

// TestValidatePickupStageTransition_ValidTransitions tests valid stage transitions
func TestValidatePickupStageTransition_ValidTransitions(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)

	validTransitions := []struct {
		name         string
		currentStage int
		newStage     int
	}{
		{"Stage 9 to 10", 9, 10},
		{"Stage 10 to 11", 10, 11},
		{"Stage 11 to 12", 11, 12},
		{"Stage 12 to 13", 12, 13},
	}

	for _, tt := range validTransitions {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePickupStageTransition(tt.currentStage, tt.newStage)
			assert.NoError(t, err, "Expected valid transition from stage %d to %d", tt.currentStage, tt.newStage)
		})
	}
}

// TestValidatePickupStageTransition_InvalidBackwardTransitions tests backward transitions are rejected
func TestValidatePickupStageTransition_InvalidBackwardTransitions(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)

	invalidTransitions := []struct {
		name         string
		currentStage int
		newStage     int
	}{
		{"Stage 11 to 10", 11, 10},
		{"Stage 12 to 11", 12, 11},
		{"Stage 13 to 12", 13, 12},
		{"Stage 13 to 10", 13, 10},
	}

	for _, tt := range invalidTransitions {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePickupStageTransition(tt.currentStage, tt.newStage)
			assert.Error(t, err, "Expected error for backward transition from stage %d to %d", tt.currentStage, tt.newStage)
			assert.Contains(t, err.Error(), "backward transition", "Error message should mention backward transition")
		})
	}
}

// TestValidatePickupStageTransition_InvalidSkippedStages tests skipped stages are rejected
func TestValidatePickupStageTransition_InvalidSkippedStages(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)

	invalidTransitions := []struct {
		name         string
		currentStage int
		newStage     int
	}{
		{"Stage 10 to 13", 10, 13},
		{"Stage 10 to 12", 10, 12},
		{"Stage 11 to 13", 11, 13},
		{"Stage 9 to 12", 9, 12},
	}

	for _, tt := range invalidTransitions {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePickupStageTransition(tt.currentStage, tt.newStage)
			assert.Error(t, err, "Expected error for skipped stages from %d to %d", tt.currentStage, tt.newStage)
			assert.Contains(t, err.Error(), "skip", "Error message should mention skipping stages")
		})
	}
}

// TestValidatePickupStageTransition_InvalidFromStage tests transitions from invalid stages
func TestValidatePickupStageTransition_InvalidFromStage(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)

	invalidTransitions := []struct {
		name         string
		currentStage int
		newStage     int
	}{
		{"Stage 5 to 10", 5, 10},
		{"Stage 8 to 11", 8, 11},
		{"Stage 1 to 12", 1, 12},
	}

	for _, tt := range invalidTransitions {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePickupStageTransition(tt.currentStage, tt.newStage)
			assert.Error(t, err, "Expected error for transition from invalid stage %d to %d", tt.currentStage, tt.newStage)
			assert.Contains(t, err.Error(), "cannot transition to pickup stage", "Error message should mention invalid starting stage")
		})
	}
}

// TestValidatePickupStageTransition_Stage13Terminal tests stage 13 is terminal
func TestValidatePickupStageTransition_Stage13Terminal(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)

	err := service.validatePickupStageTransition(13, 14)
	assert.Error(t, err, "Expected error for transition from terminal stage 13")
	assert.Contains(t, err.Error(), "final pickup stage", "Error message should mention stage 13 is final")
}

// TestUpdateOrderStatus_ValidPickupTransition tests that valid pickup transitions work end-to-end
func TestUpdateOrderStatus_ValidPickupTransition(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)
	ctx := context.Background()

	// Create test user
	user := models.User{
		FullName: "Test User",
		Email:    "test@example.com",
		Role:     "driver",
	}
	require.NoError(t, db.Create(&user).Error)

	// Create test school
	school := models.School{
		Name:      "Test School",
		Address:   "Test Address",
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}
	require.NoError(t, db.Create(&school).Error)

	// Create delivery record at stage 9
	driverID := user.ID
	deliveryRecord := models.DeliveryRecord{
		DeliveryDate:  time.Now(),
		SchoolID:      school.ID,
		DriverID:      &driverID,
		MenuItemID:    1, // Dummy menu item ID for testing
		Portions:      10,
		CurrentStage:  9,
		CurrentStatus: "sudah_diterima_pihak_sekolah",
		OmprengCount:  10,
	}
	require.NoError(t, db.Create(&deliveryRecord).Error)

	// Test valid transition from stage 9 to 10
	err := service.UpdateOrderStatus(ctx, deliveryRecord.ID, "driver_menuju_lokasi_pengambilan", 10, user.ID, "Test transition")
	assert.NoError(t, err, "Expected successful transition from stage 9 to 10")

	// Verify the delivery record was updated
	var updated models.DeliveryRecord
	require.NoError(t, db.First(&updated, deliveryRecord.ID).Error)
	assert.Equal(t, 10, updated.CurrentStage)
	assert.Equal(t, "driver_menuju_lokasi_pengambilan", updated.CurrentStatus)
}

// TestUpdateOrderStatus_InvalidPickupTransition tests that invalid pickup transitions are rejected
func TestUpdateOrderStatus_InvalidPickupTransition(t *testing.T) {
	db := setupActivityTrackerTestDB(t)
	service := NewActivityTrackerService(db)
	ctx := context.Background()

	// Create test user
	user := models.User{
		FullName: "Test User",
		Email:    "test@example.com",
		Role:     "driver",
	}
	require.NoError(t, db.Create(&user).Error)

	// Create test school
	school := models.School{
		Name:      "Test School",
		Address:   "Test Address",
		Latitude:  -6.2088,
		Longitude: 106.8456,
	}
	require.NoError(t, db.Create(&school).Error)

	// Create delivery record at stage 10
	driverID := user.ID
	deliveryRecord := models.DeliveryRecord{
		DeliveryDate:  time.Now(),
		SchoolID:      school.ID,
		DriverID:      &driverID,
		MenuItemID:    1, // Dummy menu item ID for testing
		Portions:      10,
		CurrentStage:  10,
		CurrentStatus: "driver_menuju_lokasi_pengambilan",
		OmprengCount:  10,
	}
	require.NoError(t, db.Create(&deliveryRecord).Error)

	// Test invalid transition from stage 10 to 13 (skipping stages)
	err := service.UpdateOrderStatus(ctx, deliveryRecord.ID, "driver_tiba_di_sppg", 13, user.ID, "Test invalid transition")
	assert.Error(t, err, "Expected error for invalid transition from stage 10 to 13")
	assert.Contains(t, err.Error(), "skip", "Error message should mention skipping stages")

	// Verify the delivery record was NOT updated
	var unchanged models.DeliveryRecord
	require.NoError(t, db.First(&unchanged, deliveryRecord.ID).Error)
	assert.Equal(t, 10, unchanged.CurrentStage, "Stage should remain at 10")
	assert.Equal(t, "driver_menuju_lokasi_pengambilan", unchanged.CurrentStatus, "Status should remain unchanged")
}
