package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupDeliveryTestDB creates an in-memory SQLite database for delivery task testing
func setupDeliveryTestDB(t *testing.T) *gorm.DB {
	// Use SQLite in-memory database for property tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.School{},
		&models.Recipe{},
		&models.MenuItem{},
		&models.DeliveryRecord{},
		&models.DeliveryTask{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// cleanupDeliveryTestDB cleans up the test database
func cleanupDeliveryTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM delivery_tasks")
	db.Exec("DELETE FROM delivery_records")
	db.Exec("DELETE FROM menu_items")
	db.Exec("DELETE FROM recipes")
	db.Exec("DELETE FROM schools")
	db.Exec("DELETE FROM users")
}

// TestProperty1_FaultCondition_DisplayAllPackedOrders tests Property 1
// **Validates: Requirements 2.1, 2.3, 2.5**
//
// CRITICAL: This test MUST FAIL on unfixed code - failure confirms the bug exists
// DO NOT attempt to fix the test or the code when it fails
//
// Property 1: Fault Condition - Display All Packed Orders
// For any date selection where delivery records exist with status "selesai_dipacking",
// the GetReadyOrders function SHALL return all such records regardless of whether
// driver_id is NULL or populated, allowing the form to display all packed orders
// available for delivery task creation.
//
// EXPECTED OUTCOME: Test FAILS on unfixed code (this is correct - it proves the bug exists)
func TestProperty1_FaultCondition_DisplayAllPackedOrders(t *testing.T) {
	db := setupDeliveryTestDB(t)
	defer cleanupDeliveryTestDB(db)

	deliveryTaskService := NewDeliveryTaskService(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	properties.Property("GetReadyOrders should return all selesai_dipacking records regardless of driver_id", prop.ForAll(
		func(driverIDPresent bool, numRecords int) bool {
			// Clean up before each test
			cleanupDeliveryTestDB(db)

			// Constrain numRecords to reasonable range
			if numRecords < 1 {
				numRecords = 1
			}
			if numRecords > 5 {
				numRecords = 5
			}

			// Create test date
			testDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

			// Create a driver
			driver := &models.User{
				NIK:      "DRV001",
				Email:    "driver1@sppg.test",
				FullName: "Test Driver",
				Role:     "driver",
				IsActive: true,
			}
			if err := db.Create(driver).Error; err != nil {
				t.Logf("Failed to create driver: %v", err)
				return false
			}

			// Create schools
			var schools []models.School
			for i := 0; i < numRecords; i++ {
				school := models.School{
					Name:         fmt.Sprintf("School %d", i+1),
					Address:      fmt.Sprintf("Address %d", i+1),
					Latitude:     -6.2 + float64(i)*0.01,
					Longitude:    106.8 + float64(i)*0.01,
					StudentCount: 100 + i*10,
					Category:     "SD",
					IsActive:     true,
				}
				if err := db.Create(&school).Error; err != nil {
					t.Logf("Failed to create school: %v", err)
					return false
				}
				schools = append(schools, school)
			}

			// Create recipe
			recipe := &models.Recipe{
				Name:          "Test Menu",
				Category:      "Main Course",
				TotalCalories: 500,
				TotalProtein:  20,
				TotalCarbs:    60,
				TotalFat:      15,
				IsActive:      true,
				CreatedBy:     1,
			}
			if err := db.Create(recipe).Error; err != nil {
				t.Logf("Failed to create recipe: %v", err)
				return false
			}

			// Create menu item
			menuItem := &models.MenuItem{
				MenuPlanID: 1,
				Date:       testDate,
				RecipeID:   recipe.ID,
				Portions:   100,
			}
			if err := db.Create(menuItem).Error; err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Create delivery records with mixed driver_id values
			var expectedCount int
			for i := 0; i < numRecords; i++ {
				var driverID *uint
				if driverIDPresent && i%2 == 0 {
					// Some records have driver_id set
					driverID = &driver.ID
				}
				// All records should be counted regardless of driver_id

				deliveryRecord := &models.DeliveryRecord{
					DeliveryDate:  testDate,
					SchoolID:      schools[i].ID,
					DriverID:      driverID,
					MenuItemID:    menuItem.ID,
					Portions:      50 + i*10,
					PortionsSmall: 30 + i*5,
					PortionsLarge: 20 + i*5,
					CurrentStatus: "selesai_dipacking",
					CurrentStage:  3,
					OmprengCount:  10 + i,
				}
				if err := db.Create(deliveryRecord).Error; err != nil {
					t.Logf("Failed to create delivery record: %v", err)
					return false
				}
				expectedCount++
			}

			// Call GetReadyOrders
			orders, err := deliveryTaskService.GetReadyOrders(testDate)
			if err != nil {
				t.Logf("GetReadyOrders returned error: %v", err)
				return false
			}

			// CRITICAL CHECK: All records with status "selesai_dipacking" should be returned
			// regardless of driver_id value
			if len(orders) != expectedCount {
				t.Logf("COUNTEREXAMPLE FOUND: Expected %d orders, got %d orders", expectedCount, len(orders))
				t.Logf("Test configuration: driverIDPresent=%v, numRecords=%d", driverIDPresent, numRecords)
				t.Logf("This confirms the bug: GetReadyOrders filters out records with non-NULL driver_id")
				return false
			}

			// Verify all returned orders have correct status
			for _, order := range orders {
				if order.CurrentStatus != "selesai_dipacking" {
					t.Logf("Order has incorrect status: %s", order.CurrentStatus)
					return false
				}
			}

			return true
		},
		gen.Bool(),
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}

// TestProperty2_Preservation_EmptyDataHandling tests Property 2
// **Validates: Requirements 3.1, 3.2, 3.3**
//
// Property 2: Preservation - Empty Data Handling
// For any date selection where NO delivery records exist with status "selesai_dipacking"
// OR NO active drivers exist, the fixed code SHALL produce exactly the same behavior as
// the original code, preserving the display of appropriate warning messages.
//
// EXPECTED OUTCOME: Tests PASS on unfixed code (confirms baseline behavior to preserve)
func TestProperty2_Preservation_EmptyDataHandling(t *testing.T) {
	db := setupDeliveryTestDB(t)
	defer cleanupDeliveryTestDB(db)

	deliveryTaskService := NewDeliveryTaskService(db)

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Sub-property 2.1: Dates with no delivery records should return empty array
	properties.Property("GetReadyOrders returns empty array for dates with no delivery records", prop.ForAll(
		func(dayOffset int) bool {
			// Clean up before each test
			cleanupDeliveryTestDB(db)

			// Generate a random date
			testDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, dayOffset%365)

			// Do NOT create any delivery records for this date

			// Call GetReadyOrders
			orders, err := deliveryTaskService.GetReadyOrders(testDate)
			if err != nil {
				t.Logf("GetReadyOrders returned error: %v", err)
				return false
			}

			// Should return empty array
			if len(orders) != 0 {
				t.Logf("Expected empty array for date with no records, got %d orders", len(orders))
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	// Sub-property 2.2: Delivery records with status other than "selesai_dipacking" should not be returned
	properties.Property("GetReadyOrders excludes records with status != selesai_dipacking", prop.ForAll(
		func(statusIndex int) bool {
			// Clean up before each test
			cleanupDeliveryTestDB(db)

			// Define statuses other than "selesai_dipacking"
			otherStatuses := []string{
				"pending",
				"sedang_dipacking",
				"dalam_perjalanan",
				"selesai",
				"dibatalkan",
			}

			// Pick a status from the list
			status := otherStatuses[statusIndex%len(otherStatuses)]

			testDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

			// Create school
			school := &models.School{
				Name:         "Test School",
				Address:      "Test Address",
				Latitude:     -6.2,
				Longitude:    106.8,
				StudentCount: 100,
				Category:     "SD",
				IsActive:     true,
			}
			if err := db.Create(school).Error; err != nil {
				t.Logf("Failed to create school: %v", err)
				return false
			}

			// Create recipe
			recipe := &models.Recipe{
				Name:          "Test Menu",
				Category:      "Main Course",
				TotalCalories: 500,
				TotalProtein:  20,
				TotalCarbs:    60,
				TotalFat:      15,
				IsActive:      true,
				CreatedBy:     1,
			}
			if err := db.Create(recipe).Error; err != nil {
				t.Logf("Failed to create recipe: %v", err)
				return false
			}

			// Create menu item
			menuItem := &models.MenuItem{
				MenuPlanID: 1,
				Date:       testDate,
				RecipeID:   recipe.ID,
				Portions:   100,
			}
			if err := db.Create(menuItem).Error; err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Create delivery record with status OTHER than "selesai_dipacking"
			deliveryRecord := &models.DeliveryRecord{
				DeliveryDate:  testDate,
				SchoolID:      school.ID,
				MenuItemID:    menuItem.ID,
				Portions:      50,
				PortionsSmall: 30,
				PortionsLarge: 20,
				CurrentStatus: status,
				CurrentStage:  1,
				OmprengCount:  10,
			}
			if err := db.Create(deliveryRecord).Error; err != nil {
				t.Logf("Failed to create delivery record: %v", err)
				return false
			}

			// Call GetReadyOrders
			orders, err := deliveryTaskService.GetReadyOrders(testDate)
			if err != nil {
				t.Logf("GetReadyOrders returned error: %v", err)
				return false
			}

			// Should return empty array (record should be excluded)
			if len(orders) != 0 {
				t.Logf("Expected empty array for status '%s', got %d orders", status, len(orders))
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}
