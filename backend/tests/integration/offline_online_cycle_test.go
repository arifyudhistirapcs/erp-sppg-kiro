package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type OfflineOnlineCycleTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	token  string
}

func (suite *OfflineOnlineCycleTestSuite) SetupSuite() {
	// Setup test database
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test_offline_online",
		DBSSLMode:  "disable",
		JWTSecret:  "test-secret",
	}
	
	db, err := database.Initialize(cfg)
	suite.Require().NoError(err)
	suite.db = db
	
	// Setup router
	suite.router = router.Setup(db, nil, cfg)
	
	// Create test user and get token
	suite.setupTestUser()
}

func (suite *OfflineOnlineCycleTestSuite) TearDownSuite() {
	// Clean up test database
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *OfflineOnlineCycleTestSuite) SetupTest() {
	// Clean up data before each test
	suite.db.Exec("DELETE FROM electronic_pods")
	suite.db.Exec("DELETE FROM delivery_menu_items")
	suite.db.Exec("DELETE FROM delivery_tasks")
	suite.db.Exec("DELETE FROM attendance")
	suite.db.Exec("DELETE FROM ompreng_tracking")
	suite.db.Exec("DELETE FROM schools")
}

func (suite *OfflineOnlineCycleTestSuite) setupTestUser() {
	// Create test driver user
	driver := &models.User{
		NIK:          "1234567890",
		Email:        "driver@example.com",
		PasswordHash: "$2a$10$test.hash",
		FullName:     "Driver Test",
		Role:         "driver",
		IsActive:     true,
	}
	suite.db.Create(driver)
	
	// Create employee record for driver
	employee := &models.Employee{
		UserID:      driver.ID,
		NIK:         driver.NIK,
		FullName:    driver.FullName,
		Email:       driver.Email,
		PhoneNumber: "081234567890",
		Position:    "Driver",
		JoinDate:    time.Now(),
		IsActive:    true,
	}
	suite.db.Create(employee)
	
	// Login to get token
	loginData := map[string]string{
		"email":    "driver@example.com",
		"password": "password",
	}
	
	body, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.token = response["token"].(string)
}

func (suite *OfflineOnlineCycleTestSuite) makeAuthenticatedRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}
	
	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *OfflineOnlineCycleTestSuite) TestCompleteOfflineOnlineCycle() {
	// Test PWA offline data capture → Sync → Backend verification
	
	// Step 1: Setup master data
	schoolID, deliveryTaskID := suite.setupMasterData()
	
	// Step 2: Simulate offline data capture (e-POD)
	offlineEPODData := suite.simulateOfflineEPODCapture(deliveryTaskID, schoolID)
	
	// Step 3: Simulate sync when coming back online
	suite.syncOfflineData(offlineEPODData)
	
	// Step 4: Verify backend data consistency
	suite.verifyBackendDataConsistency(deliveryTaskID, schoolID)
	
	// Step 5: Test attendance offline-online cycle
	suite.testAttendanceOfflineOnlineCycle()
	
	// Step 6: Test ompreng tracking offline-online cycle
	suite.testOmprengTrackingOfflineOnlineCycle(schoolID)
}

func (suite *OfflineOnlineCycleTestSuite) setupMasterData() (uint, uint) {
	// Create school
	school := &models.School{
		Name:          "SDN 01 Jakarta",
		Address:       "Jl. Pendidikan No. 1",
		Latitude:      -6.2088,
		Longitude:     106.8456,
		ContactPerson: "Kepala Sekolah",
		PhoneNumber:   "021-1234567",
		StudentCount:  200,
		IsActive:      true,
	}
	suite.db.Create(school)
	
	// Create delivery task
	deliveryTask := &models.DeliveryTask{
		TaskDate:   time.Now(),
		DriverID:   1, // Driver user ID
		SchoolID:   school.ID,
		Portions:   200,
		Status:     "in_progress",
		RouteOrder: 1,
	}
	suite.db.Create(deliveryTask)
	
	return school.ID, deliveryTask.ID
}

func (suite *OfflineOnlineCycleTestSuite) simulateOfflineEPODCapture(deliveryTaskID, schoolID uint) map[string]interface{} {
	// Simulate data that would be captured offline in PWA
	offlineTimestamp := time.Now()
	
	offlineEPODData := map[string]interface{}{
		"delivery_task_id": deliveryTaskID,
		"photo_url":        "blob://offline-photo-123", // Simulated offline blob URL
		"signature_url":    "blob://offline-signature-123", // Simulated offline blob URL
		"latitude":         -6.2088,
		"longitude":        106.8456,
		"recipient_name":   "Kepala Sekolah",
		"ompreng_drop_off": 10,
		"ompreng_pick_up":  8,
		"completed_at":     offlineTimestamp.Format(time.RFC3339),
		"offline_captured": true,
		"sync_status":      "pending",
	}
	
	return offlineEPODData
}

func (suite *OfflineOnlineCycleTestSuite) syncOfflineData(offlineEPODData map[string]interface{}) {
	// Simulate PWA syncing offline data when connection is restored
	
	// Convert offline blob URLs to actual storage URLs (simulated)
	offlineEPODData["photo_url"] = "https://storage.example.com/photos/photo-123.jpg"
	offlineEPODData["signature_url"] = "https://storage.example.com/signatures/signature-123.jpg"
	offlineEPODData["sync_status"] = "syncing"
	
	// Remove offline-specific fields
	delete(offlineEPODData, "offline_captured")
	delete(offlineEPODData, "sync_status")
	
	// Submit e-POD data to backend
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/epod", offlineEPODData)
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.NotEmpty(response["id"])
}

func (suite *OfflineOnlineCycleTestSuite) verifyBackendDataConsistency(deliveryTaskID, schoolID uint) {
	// Verify e-POD was created correctly
	var epod models.ElectronicPOD
	suite.db.Where("delivery_task_id = ?", deliveryTaskID).First(&epod)
	
	suite.Equal(deliveryTaskID, epod.DeliveryTaskID)
	suite.Equal("https://storage.example.com/photos/photo-123.jpg", epod.PhotoURL)
	suite.Equal("https://storage.example.com/signatures/signature-123.jpg", epod.SignatureURL)
	suite.Equal(-6.2088, epod.Latitude)
	suite.Equal(106.8456, epod.Longitude)
	suite.Equal("Kepala Sekolah", epod.RecipientName)
	suite.Equal(10, epod.OmprengDropOff)
	suite.Equal(8, epod.OmprengPickUp)
	
	// Verify delivery task status updated
	var deliveryTask models.DeliveryTask
	suite.db.First(&deliveryTask, deliveryTaskID)
	suite.Equal("completed", deliveryTask.Status)
	
	// Verify ompreng tracking created
	var omprengTracking models.OmprengTracking
	suite.db.Where("school_id = ?", schoolID).First(&omprengTracking)
	suite.Equal(10, omprengTracking.DropOff)
	suite.Equal(8, omprengTracking.PickUp)
	suite.Equal(2, omprengTracking.Balance) // 10 - 8 = 2
}

func (suite *OfflineOnlineCycleTestSuite) testAttendanceOfflineOnlineCycle() {
	// Simulate offline attendance capture
	offlineAttendanceData := map[string]interface{}{
		"employee_id": 1,
		"date":        time.Now().Format("2006-01-02"),
		"check_in":    time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
		"ssid":        "SPPG-Office-WiFi",
		"bssid":       "00:11:22:33:44:55",
		"offline_captured": true,
	}
	
	// Simulate sync when online
	delete(offlineAttendanceData, "offline_captured")
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/attendance/check-in", offlineAttendanceData)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Verify attendance record created
	var attendance models.Attendance
	suite.db.Where("employee_id = ?", 1).First(&attendance)
	suite.Equal(uint(1), attendance.EmployeeID)
	suite.Equal("SPPG-Office-WiFi", attendance.SSID)
	suite.Equal("00:11:22:33:44:55", attendance.BSSID)
	
	// Test check-out sync
	checkOutData := map[string]interface{}{
		"employee_id": 1,
		"date":        time.Now().Format("2006-01-02"),
		"check_out":   time.Now().Format(time.RFC3339),
	}
	
	w = suite.makeAuthenticatedRequest("POST", "/api/v1/attendance/check-out", checkOutData)
	suite.Equal(http.StatusOK, w.Code)
	
	// Verify work hours calculated
	suite.db.Where("employee_id = ?", 1).First(&attendance)
	suite.NotNil(attendance.CheckOut)
	suite.Greater(attendance.WorkHours, float64(7)) // Should be around 8 hours
}

func (suite *OfflineOnlineCycleTestSuite) testOmprengTrackingOfflineOnlineCycle(schoolID uint) {
	// Simulate offline ompreng tracking data
	offlineOmprengData := []map[string]interface{}{
		{
			"school_id":  schoolID,
			"date":       time.Now().Format("2006-01-02"),
			"drop_off":   5,
			"pick_up":    3,
			"recorded_by": 1,
			"offline_captured": true,
		},
		{
			"school_id":  schoolID,
			"date":       time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			"drop_off":   8,
			"pick_up":    6,
			"recorded_by": 1,
			"offline_captured": true,
		},
	}
	
	// Simulate batch sync of offline ompreng data
	for _, data := range offlineOmprengData {
		delete(data, "offline_captured")
		
		w := suite.makeAuthenticatedRequest("POST", "/api/v1/ompreng/drop-off", data)
		suite.Equal(http.StatusCreated, w.Code)
		
		w = suite.makeAuthenticatedRequest("POST", "/api/v1/ompreng/pick-up", data)
		suite.Equal(http.StatusOK, w.Code)
	}
	
	// Verify ompreng tracking records created
	var trackingRecords []models.OmprengTracking
	suite.db.Where("school_id = ?", schoolID).Find(&trackingRecords)
	suite.GreaterOrEqual(len(trackingRecords), 2) // At least 2 records from sync
	
	// Verify cumulative balance calculation
	var totalBalance int
	for _, record := range trackingRecords {
		totalBalance += record.Balance
	}
	suite.Greater(totalBalance, 0) // Should have positive balance
}

func (suite *OfflineOnlineCycleTestSuite) TestConflictResolution() {
	// Test conflict resolution when same data is modified offline and online
	
	_, deliveryTaskID := suite.setupMasterData()
	
	// Simulate online modification of delivery task
	onlineUpdateData := map[string]interface{}{
		"status": "in_progress",
		"notes":  "Updated online",
	}
	
	url := fmt.Sprintf("/api/v1/delivery-tasks/%d/status", deliveryTaskID)
	w := suite.makeAuthenticatedRequest("PUT", url, onlineUpdateData)
	suite.Equal(http.StatusOK, w.Code)
	
	// Simulate offline e-POD completion (conflicting with online status)
	offlineEPODData := map[string]interface{}{
		"delivery_task_id": deliveryTaskID,
		"photo_url":        "https://storage.example.com/photos/offline-photo.jpg",
		"signature_url":    "https://storage.example.com/signatures/offline-signature.jpg",
		"latitude":         -6.2088,
		"longitude":        106.8456,
		"recipient_name":   "Kepala Sekolah",
		"ompreng_drop_off": 12,
		"ompreng_pick_up":  10,
		"completed_at":     time.Now().Format(time.RFC3339),
	}
	
	// Sync offline data (should resolve conflict - server data wins)
	w = suite.makeAuthenticatedRequest("POST", "/api/v1/epod", offlineEPODData)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Verify conflict resolution - e-POD should be created and task status should be "completed"
	var deliveryTask models.DeliveryTask
	suite.db.First(&deliveryTask, deliveryTaskID)
	suite.Equal("completed", deliveryTask.Status) // e-POD completion should override online status
	
	var epod models.ElectronicPOD
	suite.db.Where("delivery_task_id = ?", deliveryTaskID).First(&epod)
	suite.NotEmpty(epod.PhotoURL)
	suite.NotEmpty(epod.SignatureURL)
}

func (suite *OfflineOnlineCycleTestSuite) TestPartialSyncFailure() {
	// Test handling of partial sync failures
	
	_, deliveryTaskID := suite.setupMasterData()
	
	// Simulate multiple offline operations
	offlineOperations := []map[string]interface{}{
		{
			"type": "epod",
			"data": map[string]interface{}{
				"delivery_task_id": deliveryTaskID,
				"photo_url":        "https://storage.example.com/photos/photo1.jpg",
				"signature_url":    "https://storage.example.com/signatures/signature1.jpg",
				"latitude":         -6.2088,
				"longitude":        106.8456,
				"recipient_name":   "Kepala Sekolah",
				"ompreng_drop_off": 10,
				"ompreng_pick_up":  8,
				"completed_at":     time.Now().Format(time.RFC3339),
			},
		},
		{
			"type": "attendance",
			"data": map[string]interface{}{
				"employee_id": 1,
				"date":        time.Now().Format("2006-01-02"),
				"check_in":    time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
				"ssid":        "SPPG-Office-WiFi",
				"bssid":       "00:11:22:33:44:55",
			},
		},
	}
	
	// Sync operations one by one
	successCount := 0
	for _, operation := range offlineOperations {
		var w *httptest.ResponseRecorder
		
		switch operation["type"] {
		case "epod":
			w = suite.makeAuthenticatedRequest("POST", "/api/v1/epod", operation["data"])
		case "attendance":
			w = suite.makeAuthenticatedRequest("POST", "/api/v1/attendance/check-in", operation["data"])
		}
		
		if w.Code >= 200 && w.Code < 300 {
			successCount++
		}
	}
	
	// Verify at least some operations succeeded
	suite.Greater(successCount, 0)
	
	// Verify successful operations are reflected in database
	var epod models.ElectronicPOD
	err := suite.db.Where("delivery_task_id = ?", deliveryTaskID).First(&epod).Error
	if err == nil {
		suite.NotEmpty(epod.PhotoURL)
	}
	
	var attendance models.Attendance
	err = suite.db.Where("employee_id = ?", 1).First(&attendance).Error
	if err == nil {
		suite.Equal("SPPG-Office-WiFi", attendance.SSID)
	}
}

func (suite *OfflineOnlineCycleTestSuite) TestDataIntegrityDuringSync() {
	// Test that data integrity is maintained during sync process
	
	schoolID, deliveryTaskID := suite.setupMasterData()
	
	// Create initial ompreng balance
	initialTracking := &models.OmprengTracking{
		SchoolID:   schoolID,
		Date:       time.Now().AddDate(0, 0, -1),
		DropOff:    15,
		PickUp:     10,
		Balance:    5,
		RecordedBy: 1,
	}
	suite.db.Create(initialTracking)
	
	// Simulate offline e-POD with ompreng data
	offlineEPODData := map[string]interface{}{
		"delivery_task_id": deliveryTaskID,
		"photo_url":        "https://storage.example.com/photos/integrity-test.jpg",
		"signature_url":    "https://storage.example.com/signatures/integrity-test.jpg",
		"latitude":         -6.2088,
		"longitude":        106.8456,
		"recipient_name":   "Kepala Sekolah",
		"ompreng_drop_off": 12,
		"ompreng_pick_up":  7,
		"completed_at":     time.Now().Format(time.RFC3339),
	}
	
	// Sync data
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/epod", offlineEPODData)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Verify data integrity
	// 1. e-POD should be created
	var epod models.ElectronicPOD
	suite.db.Where("delivery_task_id = ?", deliveryTaskID).First(&epod)
	suite.Equal(12, epod.OmprengDropOff)
	suite.Equal(7, epod.OmprengPickUp)
	
	// 2. Ompreng tracking should be updated
	var newTracking models.OmprengTracking
	suite.db.Where("school_id = ? AND date = ?", schoolID, time.Now().Format("2006-01-02")).First(&newTracking)
	suite.Equal(12, newTracking.DropOff)
	suite.Equal(7, newTracking.PickUp)
	suite.Equal(5, newTracking.Balance) // 12 - 7 = 5
	
	// 3. Global ompreng inventory should be consistent
	var totalDropOff, totalPickUp int
	suite.db.Model(&models.OmprengTracking{}).Where("school_id = ?", schoolID).Select("SUM(drop_off)").Scan(&totalDropOff)
	suite.db.Model(&models.OmprengTracking{}).Where("school_id = ?", schoolID).Select("SUM(pick_up)").Scan(&totalPickUp)
	
	expectedBalance := totalDropOff - totalPickUp
	
	var actualBalance int
	suite.db.Model(&models.OmprengTracking{}).Where("school_id = ?", schoolID).Select("SUM(balance)").Scan(&actualBalance)
	
	suite.Equal(expectedBalance, actualBalance)
}

func (suite *OfflineOnlineCycleTestSuite) TestSyncStatusTracking() {
	// Test that sync status is properly tracked and reported
	
	_, deliveryTaskID := suite.setupMasterData()
	
	// Simulate offline data with sync metadata
	offlineDataWithMetadata := map[string]interface{}{
		"delivery_task_id": deliveryTaskID,
		"photo_url":        "https://storage.example.com/photos/sync-status-test.jpg",
		"signature_url":    "https://storage.example.com/signatures/sync-status-test.jpg",
		"latitude":         -6.2088,
		"longitude":        106.8456,
		"recipient_name":   "Kepala Sekolah",
		"ompreng_drop_off": 8,
		"ompreng_pick_up":  6,
		"completed_at":     time.Now().Format(time.RFC3339),
		"client_timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339), // Captured 1 hour ago offline
		"sync_attempt":     1,
	}
	
	// Sync data
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/epod", offlineDataWithMetadata)
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	// Verify sync was successful
	suite.NotEmpty(response["id"])
	
	// Verify server timestamp vs client timestamp handling
	var epod models.ElectronicPOD
	suite.db.Where("delivery_task_id = ?", deliveryTaskID).First(&epod)
	
	// Server should use its own timestamp for completed_at, not client timestamp
	suite.WithinDuration(time.Now(), epod.CompletedAt, 5*time.Second)
}

func TestOfflineOnlineCycleTestSuite(t *testing.T) {
	suite.Run(t, new(OfflineOnlineCycleTestSuite))
}