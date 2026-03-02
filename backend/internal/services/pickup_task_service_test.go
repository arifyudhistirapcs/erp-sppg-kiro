package services

import (
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPickupTaskTestDB creates an in-memory SQLite database for testing
func setupPickupTaskTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.User{},
		&models.School{},
		&models.DeliveryRecord{},
		&models.PickupTask{},
		&models.StatusTransition{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// cleanupPickupTaskTestDB closes the database connection
func cleanupPickupTaskTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// createTestDriver creates a test driver user
func createTestDriver(t *testing.T, db *gorm.DB) *models.User {
	driver := &models.User{
		FullName:    "Test Driver",
		Role:        "driver",
		PhoneNumber: "081234567890",
		IsActive:    true,
	}
	err := db.Create(driver).Error
	require.NoError(t, err, "Failed to create test driver")
	return driver
}

// createTestPickupSchool creates a test school for pickup task tests
func createTestPickupSchool(t *testing.T, db *gorm.DB, name string) *models.School {
	school := &models.School{
		Name:         name,
		Address:      "Test Address",
		Latitude:     -6.2088,
		Longitude:    106.8456,
		Category:     "SD",
		StudentCount: 100,
		IsActive:     true,
	}
	err := db.Create(school).Error
	require.NoError(t, err, "Failed to create test school")
	return school
}

// createTestDeliveryRecord creates a test delivery record
func createTestDeliveryRecord(t *testing.T, db *gorm.DB, schoolID uint, stage int, status string) *models.DeliveryRecord {
	deliveryRecord := &models.DeliveryRecord{
		DeliveryDate:  time.Now(),
		SchoolID:      schoolID,
		CurrentStage:  stage,
		CurrentStatus: status,
		OmprengCount:  15,
	}
	err := db.Create(deliveryRecord).Error
	require.NoError(t, err, "Failed to create test delivery record")
	return deliveryRecord
}

// createTestPickupTask creates a test pickup task with delivery records
func createTestPickupTask(t *testing.T, db *gorm.DB, driverID uint, deliveryRecordIDs []uint) *models.PickupTask {
	pickupTask := &models.PickupTask{
		TaskDate: time.Now(),
		DriverID: driverID,
		Status:   "active",
	}
	err := db.Create(pickupTask).Error
	require.NoError(t, err, "Failed to create test pickup task")

	// Update delivery records with pickup task ID and route order
	for i, drID := range deliveryRecordIDs {
		err := db.Model(&models.DeliveryRecord{}).
			Where("id = ?", drID).
			Updates(map[string]interface{}{
				"pickup_task_id": pickupTask.ID,
				"route_order":    i + 1,
			}).Error
		require.NoError(t, err, "Failed to update delivery record with pickup task")
	}

	return pickupTask
}

// TestUpdateDeliveryRecordStage_Success tests successful stage transition from 10 to 11
func TestUpdateDeliveryRecordStage_Success(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data
	driver := createTestDriver(t, db)
	school := createTestPickupSchool(t, db, "SD Test 1")
	deliveryRecord := createTestDeliveryRecord(t, db, school.ID, 10, "driver_menuju_lokasi_pengambilan")
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{deliveryRecord.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Update stage from 10 to 11
	result, err := service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 11, "driver_tiba_di_lokasi_pengambilan")

	// Assertions
	require.NoError(t, err, "Expected no error")
	require.NotNil(t, result, "Expected result to be non-nil")
	assert.Equal(t, 11, result.CurrentStage, "Expected stage to be 11")
	assert.Equal(t, "driver_tiba_di_lokasi_pengambilan", result.CurrentStatus, "Expected status to match")

	// Verify status transition was recorded
	var transition models.StatusTransition
	err = db.Where("delivery_record_id = ? AND stage = ?", deliveryRecord.ID, 11).First(&transition).Error
	require.NoError(t, err, "Expected status transition to be recorded")
	assert.Equal(t, "driver_menuju_lokasi_pengambilan", transition.FromStatus)
	assert.Equal(t, "driver_tiba_di_lokasi_pengambilan", transition.ToStatus)
}

// TestUpdateDeliveryRecordStage_SequentialTransitions tests multiple sequential stage transitions
func TestUpdateDeliveryRecordStage_SequentialTransitions(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data
	driver := createTestDriver(t, db)
	school := createTestPickupSchool(t, db, "SD Test 1")
	deliveryRecord := createTestDeliveryRecord(t, db, school.ID, 10, "driver_menuju_lokasi_pengambilan")
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{deliveryRecord.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Transition 10 -> 11
	result, err := service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 11, "driver_tiba_di_lokasi_pengambilan")
	require.NoError(t, err)
	assert.Equal(t, 11, result.CurrentStage)

	// Transition 11 -> 12
	result, err = service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 12, "driver_kembali_ke_sppg")
	require.NoError(t, err)
	assert.Equal(t, 12, result.CurrentStage)

	// Transition 12 -> 13
	result, err = service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 13, "driver_tiba_di_sppg")
	require.NoError(t, err)
	assert.Equal(t, 13, result.CurrentStage)

	// Verify pickup task status is still active (only one delivery record)
	var updatedTask models.PickupTask
	err = db.Where("id = ?", pickupTask.ID).First(&updatedTask).Error
	require.NoError(t, err)
	assert.Equal(t, "completed", updatedTask.Status, "Expected pickup task to be completed when all records at stage 13")
}

// TestUpdateDeliveryRecordStage_SkipStage tests that skipping stages is rejected
func TestUpdateDeliveryRecordStage_SkipStage(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data
	driver := createTestDriver(t, db)
	school := createTestPickupSchool(t, db, "SD Test 1")
	deliveryRecord := createTestDeliveryRecord(t, db, school.ID, 10, "driver_menuju_lokasi_pengambilan")
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{deliveryRecord.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Attempt to skip from stage 10 to 12
	result, err := service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 12, "driver_kembali_ke_sppg")

	// Assertions
	require.Error(t, err, "Expected error when skipping stages")
	assert.Nil(t, result, "Expected result to be nil")
	assert.Contains(t, err.Error(), "cannot skip stages", "Expected error message about skipping stages")
	assert.Contains(t, err.Error(), "Current stage is 10", "Expected error to mention current stage")
	assert.Contains(t, err.Error(), "attempted stage is 12", "Expected error to mention attempted stage")

	// Verify delivery record stage was not changed
	var dr models.DeliveryRecord
	err = db.Where("id = ?", deliveryRecord.ID).First(&dr).Error
	require.NoError(t, err)
	assert.Equal(t, 10, dr.CurrentStage, "Expected stage to remain at 10")
}

// TestUpdateDeliveryRecordStage_InvalidStageStatusMapping tests invalid stage-status combinations
func TestUpdateDeliveryRecordStage_InvalidStageStatusMapping(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data
	driver := createTestDriver(t, db)
	school := createTestPickupSchool(t, db, "SD Test 1")
	deliveryRecord := createTestDeliveryRecord(t, db, school.ID, 10, "driver_menuju_lokasi_pengambilan")
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{deliveryRecord.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Attempt to use wrong status for stage 11
	result, err := service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 11, "driver_kembali_ke_sppg")

	// Assertions
	require.Error(t, err, "Expected error for invalid stage-status mapping")
	assert.Nil(t, result, "Expected result to be nil")
	assert.Contains(t, err.Error(), "invalid status for stage 11", "Expected error about invalid status")
}

// TestUpdateDeliveryRecordStage_DeliveryRecordNotInTask tests updating a delivery record not in the pickup task
func TestUpdateDeliveryRecordStage_DeliveryRecordNotInTask(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data
	driver := createTestDriver(t, db)
	school1 := createTestPickupSchool(t, db, "SD Test 1")
	school2 := createTestPickupSchool(t, db, "SD Test 2")
	deliveryRecord1 := createTestDeliveryRecord(t, db, school1.ID, 10, "driver_menuju_lokasi_pengambilan")
	deliveryRecord2 := createTestDeliveryRecord(t, db, school2.ID, 10, "driver_menuju_lokasi_pengambilan")
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{deliveryRecord1.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Attempt to update delivery record 2 which is not in the pickup task
	result, err := service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord2.ID, 11, "driver_tiba_di_lokasi_pengambilan")

	// Assertions
	require.Error(t, err, "Expected error when delivery record not in task")
	assert.Nil(t, result, "Expected result to be nil")
	assert.Contains(t, err.Error(), "not part of pickup task", "Expected error about delivery record not in task")
}

// TestUpdateDeliveryRecordStage_FinalStage tests that stage 13 cannot be updated further
func TestUpdateDeliveryRecordStage_FinalStage(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data
	driver := createTestDriver(t, db)
	school := createTestPickupSchool(t, db, "SD Test 1")
	deliveryRecord := createTestDeliveryRecord(t, db, school.ID, 13, "driver_tiba_di_sppg")
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{deliveryRecord.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Attempt to update from stage 13 (should fail because stage must be between 10 and 12)
	// We'll try to transition to stage 14 which is invalid
	result, err := service.UpdateDeliveryRecordStage(pickupTask.ID, deliveryRecord.ID, 14, "some_status")

	// Assertions
	require.Error(t, err, "Expected error when trying to update final stage")
	assert.Nil(t, result, "Expected result to be nil")
	// The error could be either about invalid stage (14) or about current stage being 13
	// Both are valid error messages
	assert.True(t, 
		err.Error() == "invalid stage: 14. Must be 11, 12, or 13" || 
		err.Error() == "cannot update stage: current stage is 13 (must be between 10 and 12)",
		"Expected error about invalid stage or cannot update stage 13")
}

// TestUpdateDeliveryRecordStage_MultipleRecordsAutoComplete tests automatic completion when all records reach stage 13
func TestUpdateDeliveryRecordStage_MultipleRecordsAutoComplete(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data with 3 delivery records
	driver := createTestDriver(t, db)
	school1 := createTestPickupSchool(t, db, "SD Test 1")
	school2 := createTestPickupSchool(t, db, "SD Test 2")
	school3 := createTestPickupSchool(t, db, "SD Test 3")
	
	dr1 := createTestDeliveryRecord(t, db, school1.ID, 10, "driver_menuju_lokasi_pengambilan")
	dr2 := createTestDeliveryRecord(t, db, school2.ID, 10, "driver_menuju_lokasi_pengambilan")
	dr3 := createTestDeliveryRecord(t, db, school3.ID, 10, "driver_menuju_lokasi_pengambilan")
	
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{dr1.ID, dr2.ID, dr3.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Transition all records to stage 13
	// Record 1: 10 -> 11 -> 12 -> 13
	_, err := service.UpdateDeliveryRecordStage(pickupTask.ID, dr1.ID, 11, "driver_tiba_di_lokasi_pengambilan")
	require.NoError(t, err)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr1.ID, 12, "driver_kembali_ke_sppg")
	require.NoError(t, err)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr1.ID, 13, "driver_tiba_di_sppg")
	require.NoError(t, err)

	// Verify pickup task is still active (not all records at stage 13)
	var task1 models.PickupTask
	err = db.Where("id = ?", pickupTask.ID).First(&task1).Error
	require.NoError(t, err)
	assert.Equal(t, "active", task1.Status, "Expected pickup task to still be active")

	// Record 2: 10 -> 11 -> 12 -> 13
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr2.ID, 11, "driver_tiba_di_lokasi_pengambilan")
	require.NoError(t, err)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr2.ID, 12, "driver_kembali_ke_sppg")
	require.NoError(t, err)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr2.ID, 13, "driver_tiba_di_sppg")
	require.NoError(t, err)

	// Verify pickup task is still active
	var task2 models.PickupTask
	err = db.Where("id = ?", pickupTask.ID).First(&task2).Error
	require.NoError(t, err)
	assert.Equal(t, "active", task2.Status, "Expected pickup task to still be active")

	// Record 3: 10 -> 11 -> 12 -> 13 (this should trigger auto-completion)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr3.ID, 11, "driver_tiba_di_lokasi_pengambilan")
	require.NoError(t, err)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr3.ID, 12, "driver_kembali_ke_sppg")
	require.NoError(t, err)
	_, err = service.UpdateDeliveryRecordStage(pickupTask.ID, dr3.ID, 13, "driver_tiba_di_sppg")
	require.NoError(t, err)

	// Verify pickup task is now completed
	var task3 models.PickupTask
	err = db.Where("id = ?", pickupTask.ID).First(&task3).Error
	require.NoError(t, err)
	assert.Equal(t, "completed", task3.Status, "Expected pickup task to be completed when all records at stage 13")
}

// TestUpdateDeliveryRecordStage_IndependentStageTracking tests that updating one record doesn't affect others
func TestUpdateDeliveryRecordStage_IndependentStageTracking(t *testing.T) {
	db := setupPickupTaskTestDB(t)
	defer cleanupPickupTaskTestDB(db)

	// Create test data with 2 delivery records
	driver := createTestDriver(t, db)
	school1 := createTestPickupSchool(t, db, "SD Test 1")
	school2 := createTestPickupSchool(t, db, "SD Test 2")
	
	dr1 := createTestDeliveryRecord(t, db, school1.ID, 10, "driver_menuju_lokasi_pengambilan")
	dr2 := createTestDeliveryRecord(t, db, school2.ID, 10, "driver_menuju_lokasi_pengambilan")
	
	pickupTask := createTestPickupTask(t, db, driver.ID, []uint{dr1.ID, dr2.ID})

	// Create service
	ats := NewActivityTrackerService(db)
	service := NewPickupTaskService(db, ats)

	// Update only record 1 to stage 11
	_, err := service.UpdateDeliveryRecordStage(pickupTask.ID, dr1.ID, 11, "driver_tiba_di_lokasi_pengambilan")
	require.NoError(t, err)

	// Verify record 1 is at stage 11
	var updatedDr1 models.DeliveryRecord
	err = db.Where("id = ?", dr1.ID).First(&updatedDr1).Error
	require.NoError(t, err)
	assert.Equal(t, 11, updatedDr1.CurrentStage, "Expected record 1 to be at stage 11")

	// Verify record 2 is still at stage 10
	var updatedDr2 models.DeliveryRecord
	err = db.Where("id = ?", dr2.ID).First(&updatedDr2).Error
	require.NoError(t, err)
	assert.Equal(t, 10, updatedDr2.CurrentStage, "Expected record 2 to remain at stage 10")
}
