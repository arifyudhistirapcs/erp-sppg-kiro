package services

import (
	"context"
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

// setupKDSBugfixTestDB creates an in-memory SQLite database for KDS bugfix testing
func setupKDSBugfixTestDB(t *testing.T) *gorm.DB {
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
		&models.RecipeItem{},
		&models.SemiFinishedGoods{},
		&models.SemiFinishedInventory{},
		&models.InventoryMovement{},
		&models.MenuPlan{},
		&models.MenuItem{},
		&models.MenuItemSchoolAllocation{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// cleanupKDSBugfixTestDB cleans up the test database
func cleanupKDSBugfixTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM inventory_movements")
	db.Exec("DELETE FROM semi_finished_inventories")
	db.Exec("DELETE FROM menu_item_school_allocations")
	db.Exec("DELETE FROM menu_items")
	db.Exec("DELETE FROM menu_plans")
	db.Exec("DELETE FROM recipe_items")
	db.Exec("DELETE FROM recipes")
	db.Exec("DELETE FROM semi_finished_goods")
	db.Exec("DELETE FROM schools")
	db.Exec("DELETE FROM users")
}

// TestProperty1_BugCondition_StockValidationBeforeCooking tests the bug condition
// **Validates: Requirements 2.1, 2.2, 2.3**
// 
// CRITICAL: This test MUST FAIL on unfixed code - failure confirms the bug exists
// DO NOT attempt to fix the test or the code when it fails
// 
// This test encodes the expected behavior:
// - Stock validation SHOULD be performed before allowing status change to "cooking"
// - Detailed error message SHOULD be returned when stock is insufficient
// - Stock deduction SHOULD happen automatically when validation passes
//
// EXPECTED OUTCOME on UNFIXED code: Test FAILS (this is correct - it proves the bug exists)
// EXPECTED OUTCOME on FIXED code: Test PASSES (confirms bug is fixed)
//
// NOTE: This test focuses on the deductInventory method which is currently disabled in UpdateRecipeStatus
func TestProperty1_BugCondition_StockValidationBeforeCooking(t *testing.T) {
	db := setupKDSBugfixTestDB(t)
	defer cleanupKDSBugfixTestDB(db)

	// Create KDS service directly without Firebase (for testing)
	kdsService := &KDSService{
		db:                db,
		firebaseApp:       nil,
		dbClient:          nil,
		monitoringService: nil,
	}

	properties := gopter.NewProperties(nil)

	// Property 1: Stock validation should prevent cooking when stock is insufficient
	// This tests the deductInventory method directly, which is currently DISABLED in UpdateRecipeStatus
	properties.Property("Stock validation prevents cooking with insufficient stock", prop.ForAll(
		func(smallPortions, largePortions int, stockMultiplier float64) bool {
			cleanupKDSBugfixTestDB(db)
			ctx := context.Background()

			// Create test user
			user := models.User{
				NIK:          "1234567890",
				Email:        "chef@test.com",
				PasswordHash: "hashed_password",
				FullName:     "Test Chef",
				Role:         "chef",
				IsActive:     true,
			}
			if err := db.Create(&user).Error; err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create semi-finished goods (Nasi and Ayam Goreng)
			nasiGoods := models.SemiFinishedGoods{
				Name: "Nasi",
				Unit: "kg",
			}
			ayamGoods := models.SemiFinishedGoods{
				Name: "Ayam Goreng",
				Unit: "kg",
			}
			if err := db.Create(&nasiGoods).Error; err != nil {
				t.Logf("Failed to create nasi goods: %v", err)
				return false
			}
			if err := db.Create(&ayamGoods).Error; err != nil {
				t.Logf("Failed to create ayam goods: %v", err)
				return false
			}

			// Create recipe
			recipe := models.Recipe{
				Name:          "Paket Ayam Goreng",
				Category:      "Paket",
				Instructions:  "Test recipe",
				TotalCalories: 500,
				TotalProtein:  20,
				TotalCarbs:    60,
				TotalFat:      15,
				CreatedBy:     user.ID,
			}
			if err := db.Create(&recipe).Error; err != nil {
				t.Logf("Failed to create recipe: %v", err)
				return false
			}

			// Create recipe items with portion-specific quantities
			// Small portion: 0.05 kg nasi, 0.08 kg ayam
			// Large portion: 0.10 kg nasi, 0.15 kg ayam
			nasiItem := models.RecipeItem{
				RecipeID:                recipe.ID,
				SemiFinishedGoodsID:     nasiGoods.ID,
				Quantity:                0, // deprecated field
				QuantityPerPortionSmall: 0.05,
				QuantityPerPortionLarge: 0.10,
			}
			ayamItem := models.RecipeItem{
				RecipeID:                recipe.ID,
				SemiFinishedGoodsID:     ayamGoods.ID,
				Quantity:                0, // deprecated field
				QuantityPerPortionSmall: 0.08,
				QuantityPerPortionLarge: 0.15,
			}
			if err := db.Create(&nasiItem).Error; err != nil {
				t.Logf("Failed to create nasi item: %v", err)
				return false
			}
			if err := db.Create(&ayamItem).Error; err != nil {
				t.Logf("Failed to create ayam item: %v", err)
				return false
			}

			// Create inventory with stock based on multiplier
			nasiInventory := models.SemiFinishedInventory{
				SemiFinishedGoodsID: nasiGoods.ID,
				Quantity:            10.0, // Fixed amount for testing
				LastUpdated:         time.Now(),
			}
			ayamInventory := models.SemiFinishedInventory{
				SemiFinishedGoodsID: ayamGoods.ID,
				Quantity:            10.0, // Fixed amount for testing
				LastUpdated:         time.Now(),
			}
			if err := db.Create(&nasiInventory).Error; err != nil {
				t.Logf("Failed to create nasi inventory: %v", err)
				return false
			}
			if err := db.Create(&ayamInventory).Error; err != nil {
				t.Logf("Failed to create ayam inventory: %v", err)
				return false
			}

			// Store initial stock levels
			initialNasiStock := nasiInventory.Quantity
			initialAyamStock := ayamInventory.Quantity

			// Create menu plan and menu item with school allocations
			menuPlan := models.MenuPlan{
				WeekStart: time.Now(),
				WeekEnd:   time.Now().AddDate(0, 0, 7),
				Status:    "approved",
				CreatedBy: user.ID,
			}
			if err := db.Create(&menuPlan).Error; err != nil {
				t.Logf("Failed to create menu plan: %v", err)
				return false
			}

			// Create school for allocations
			school := models.School{
				Name:         "Test School",
				Address:      "Test Address",
				Latitude:     -6.2,
				Longitude:    106.8,
				StudentCount: 100,
				Category:     "SD",
			}
			if err := db.Create(&school).Error; err != nil {
				t.Logf("Failed to create school: %v", err)
				return false
			}

			// Create menu item
			menuItem := models.MenuItem{
				MenuPlanID: menuPlan.ID,
				Date:       time.Now(),
				RecipeID:   recipe.ID,
				Portions:   smallPortions + largePortions,
			}
			if err := db.Create(&menuItem).Error; err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Create school allocations with mixed portion sizes
			if smallPortions > 0 {
				smallAlloc := models.MenuItemSchoolAllocation{
					MenuItemID:  menuItem.ID,
					SchoolID:    school.ID,
					Portions:    smallPortions,
					PortionSize: "small",
					Date:        time.Now(),
				}
				if err := db.Create(&smallAlloc).Error; err != nil {
					t.Logf("Failed to create small allocation: %v", err)
					return false
				}
				menuItem.SchoolAllocations = append(menuItem.SchoolAllocations, smallAlloc)
			}

			if largePortions > 0 {
				largeAlloc := models.MenuItemSchoolAllocation{
					MenuItemID:  menuItem.ID,
					SchoolID:    school.ID,
					Portions:    largePortions,
					PortionSize: "large",
					Date:        time.Now(),
				}
				if err := db.Create(&largeAlloc).Error; err != nil {
					t.Logf("Failed to create large allocation: %v", err)
					return false
				}
				menuItem.SchoolAllocations = append(menuItem.SchoolAllocations, largeAlloc)
			}

			// Load recipe with items
			if err := db.Preload("RecipeItems").Preload("RecipeItems.SemiFinishedGoods").First(&menuItem.Recipe, recipe.ID).Error; err != nil {
				t.Logf("Failed to load recipe: %v", err)
				return false
			}

			// Call deductInventory directly (this is what UpdateRecipeStatus should call but doesn't)
			err := kdsService.deductInventory(ctx, &menuItem, user.ID)

			// Calculate expected deductions based on portion sizes
			expectedNasiDeduction := float64(smallPortions)*0.05 + float64(largePortions)*0.10
			expectedAyamDeduction := float64(smallPortions)*0.08 + float64(largePortions)*0.15

			// EXPECTED BEHAVIOR (after fix):
			// 1. If stock is sufficient, deduction should succeed and stock should be reduced
			// 2. If stock is insufficient, error should be returned
			// 3. Deduction should be based on portion size calculations, not deprecated Quantity field

			if expectedNasiDeduction > initialNasiStock || expectedAyamDeduction > initialAyamStock {
				// Stock is insufficient - should return error
				if err == nil {
					t.Logf("COUNTEREXAMPLE: Expected error for insufficient stock, but got success")
					t.Logf("  Nasi: need %.2f kg, have %.2f kg", expectedNasiDeduction, initialNasiStock)
					t.Logf("  Ayam: need %.2f kg, have %.2f kg", expectedAyamDeduction, initialAyamStock)
					return false
				}
				// Error returned as expected
				return true
			}

			// Stock is sufficient - should succeed and deduct
			if err != nil {
				t.Logf("Unexpected error with sufficient stock: %v", err)
				t.Logf("  Nasi: need %.2f kg, have %.2f kg", expectedNasiDeduction, initialNasiStock)
				t.Logf("  Ayam: need %.2f kg, have %.2f kg", expectedAyamDeduction, initialAyamStock)
				return false
			}

			// Verify stock was deducted correctly
			var currentNasiInventory models.SemiFinishedInventory
			var currentAyamInventory models.SemiFinishedInventory
			db.Where("semi_finished_goods_id = ?", nasiGoods.ID).First(&currentNasiInventory)
			db.Where("semi_finished_goods_id = ?", ayamGoods.ID).First(&currentAyamInventory)

			// Check if stock was deducted by the expected amounts
			actualNasiDeduction := initialNasiStock - currentNasiInventory.Quantity
			actualAyamDeduction := initialAyamStock - currentAyamInventory.Quantity

			// Allow small floating point differences
			nasiMatch := actualNasiDeduction >= expectedNasiDeduction-0.001 && actualNasiDeduction <= expectedNasiDeduction+0.001
			ayamMatch := actualAyamDeduction >= expectedAyamDeduction-0.001 && actualAyamDeduction <= expectedAyamDeduction+0.001

			if !nasiMatch || !ayamMatch {
				t.Logf("COUNTEREXAMPLE: Stock deduction amounts incorrect")
				t.Logf("  Small portions: %d, Large portions: %d", smallPortions, largePortions)
				t.Logf("  Nasi: expected %.2f kg, actual %.2f kg", expectedNasiDeduction, actualNasiDeduction)
				t.Logf("  Ayam: expected %.2f kg, actual %.2f kg", expectedAyamDeduction, actualAyamDeduction)
				return false
			}

			return true // Expected behavior: correct deduction based on portion sizes
		},
		gen.IntRange(10, 100),      // smallPortions: 10-100
		gen.IntRange(10, 100),      // largePortions: 10-100
		gen.Float64Range(0.3, 1.5), // stockMultiplier: 0.3-1.5 (not used in simplified test)
	))

	properties.TestingRun(t)
}

// TestProperty2_Preservation_NonCookingStatusUpdates tests preservation of non-cooking status updates
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
//
// IMPORTANT: This test runs on UNFIXED code to observe baseline behavior
// EXPECTED OUTCOME: Test PASSES (confirms baseline behavior to preserve)
//
// This test verifies that status updates to "ready" and "pending" work correctly
// without any stock validation or deduction, and this behavior should be preserved after the fix.
func TestProperty2_Preservation_NonCookingStatusUpdates(t *testing.T) {
	db := setupKDSBugfixTestDB(t)
	defer cleanupKDSBugfixTestDB(db)

	// Create KDS service directly without Firebase (for testing)
	kdsService := &KDSService{
		db:                db,
		firebaseApp:       nil,
		dbClient:          nil,
		monitoringService: nil,
	}

	properties := gopter.NewProperties(nil)

	// Property 2: Non-cooking status updates should not perform stock validation or deduction
	properties.Property("Non-cooking status updates preserve existing behavior", prop.ForAll(
		func(statusChoice int, portions int) bool {
			cleanupKDSBugfixTestDB(db)
			ctx := context.Background()

			// Map statusChoice to non-cooking statuses
			statuses := []string{"pending", "ready"}
			status := statuses[statusChoice%len(statuses)]

			// Create test user
			user := models.User{
				NIK:          "1234567890",
				Email:        "chef@test.com",
				PasswordHash: "hashed_password",
				FullName:     "Test Chef",
				Role:         "chef",
				IsActive:     true,
			}
			if err := db.Create(&user).Error; err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create school
			school := models.School{
				Name:         "Test School",
				Address:      "Test Address",
				Latitude:     -6.2,
				Longitude:    106.8,
				StudentCount: 100,
				Category:     "SMP",
				IsActive:     true,
			}
			if err := db.Create(&school).Error; err != nil {
				t.Logf("Failed to create school: %v", err)
				return false
			}

			// Create semi-finished goods
			nasiGoods := models.SemiFinishedGoods{
				Name: "Nasi",
				Unit: "kg",
			}
			ayamGoods := models.SemiFinishedGoods{
				Name: "Ayam Goreng",
				Unit: "kg",
			}
			if err := db.Create(&nasiGoods).Error; err != nil {
				t.Logf("Failed to create nasi goods: %v", err)
				return false
			}
			if err := db.Create(&ayamGoods).Error; err != nil {
				t.Logf("Failed to create ayam goods: %v", err)
				return false
			}

			// Create recipe
			recipe := models.Recipe{
				Name:          "Paket Ayam Goreng",
				Category:      "Paket",
				Instructions:  "Test recipe",
				TotalCalories: 500,
				TotalProtein:  20,
				TotalCarbs:    60,
				TotalFat:      15,
				CreatedBy:     user.ID,
			}
			if err := db.Create(&recipe).Error; err != nil {
				t.Logf("Failed to create recipe: %v", err)
				return false
			}

			// Create recipe items
			nasiItem := models.RecipeItem{
				RecipeID:                recipe.ID,
				SemiFinishedGoodsID:     nasiGoods.ID,
				Quantity:                0,
				QuantityPerPortionSmall: 0.05,
				QuantityPerPortionLarge: 0.10,
			}
			ayamItem := models.RecipeItem{
				RecipeID:                recipe.ID,
				SemiFinishedGoodsID:     ayamGoods.ID,
				Quantity:                0,
				QuantityPerPortionSmall: 0.08,
				QuantityPerPortionLarge: 0.15,
			}
			if err := db.Create(&nasiItem).Error; err != nil {
				t.Logf("Failed to create nasi item: %v", err)
				return false
			}
			if err := db.Create(&ayamItem).Error; err != nil {
				t.Logf("Failed to create ayam item: %v", err)
				return false
			}

			// Create inventory with some stock
			nasiInventory := models.SemiFinishedInventory{
				SemiFinishedGoodsID: nasiGoods.ID,
				Quantity:            10.0,
				LastUpdated:         time.Now(),
			}
			ayamInventory := models.SemiFinishedInventory{
				SemiFinishedGoodsID: ayamGoods.ID,
				Quantity:            10.0,
				LastUpdated:         time.Now(),
			}
			if err := db.Create(&nasiInventory).Error; err != nil {
				t.Logf("Failed to create nasi inventory: %v", err)
				return false
			}
			if err := db.Create(&ayamInventory).Error; err != nil {
				t.Logf("Failed to create ayam inventory: %v", err)
				return false
			}

			// Store initial stock levels
			initialNasiStock := nasiInventory.Quantity
			initialAyamStock := ayamInventory.Quantity

			// Create menu plan
			today := time.Now()
			weekStart := today.AddDate(0, 0, -int(today.Weekday()))
			weekEnd := weekStart.AddDate(0, 0, 6)
			menuPlan := models.MenuPlan{
				WeekStart: weekStart,
				WeekEnd:   weekEnd,
				Status:    "approved",
				CreatedBy: user.ID,
			}
			if err := db.Create(&menuPlan).Error; err != nil {
				t.Logf("Failed to create menu plan: %v", err)
				return false
			}

			// Create menu item for today
			menuItem := models.MenuItem{
				MenuPlanID: menuPlan.ID,
				Date:       today,
				RecipeID:   recipe.ID,
				Portions:   portions,
			}
			if err := db.Create(&menuItem).Error; err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Create school allocation
			allocation := models.MenuItemSchoolAllocation{
				MenuItemID:  menuItem.ID,
				SchoolID:    school.ID,
				Portions:    portions,
				PortionSize: "large",
				Date:        today,
			}
			if err := db.Create(&allocation).Error; err != nil {
				t.Logf("Failed to create allocation: %v", err)
				return false
			}

			// Call UpdateRecipeStatus with non-cooking status
			err := kdsService.UpdateRecipeStatus(ctx, recipe.ID, status, user.ID)

			// EXPECTED BEHAVIOR: Status update should succeed without stock validation
			if err != nil {
				t.Logf("Unexpected error for status '%s': %v", status, err)
				return false
			}

			// Verify stock was NOT deducted (preservation requirement)
			var currentNasiInventory models.SemiFinishedInventory
			var currentAyamInventory models.SemiFinishedInventory
			db.Where("semi_finished_goods_id = ?", nasiGoods.ID).First(&currentNasiInventory)
			db.Where("semi_finished_goods_id = ?", ayamGoods.ID).First(&currentAyamInventory)

			// Stock should remain unchanged for non-cooking status updates
			if currentNasiInventory.Quantity != initialNasiStock {
				t.Logf("PRESERVATION VIOLATION: Nasi stock changed for status '%s' (was %.2f, now %.2f)",
					status, initialNasiStock, currentNasiInventory.Quantity)
				return false
			}
			if currentAyamInventory.Quantity != initialAyamStock {
				t.Logf("PRESERVATION VIOLATION: Ayam stock changed for status '%s' (was %.2f, now %.2f)",
					status, initialAyamStock, currentAyamInventory.Quantity)
				return false
			}

			// Verify no inventory movements were created
			var movementCount int64
			db.Model(&models.InventoryMovement{}).Count(&movementCount)
			if movementCount > 0 {
				t.Logf("PRESERVATION VIOLATION: Inventory movements created for status '%s'", status)
				return false
			}

			return true
		},
		gen.IntRange(0, 1),    // statusChoice: 0 (pending) or 1 (ready)
		gen.IntRange(10, 100), // portions: 10-100
	))

	properties.TestingRun(t)
}

// TestProperty3_Preservation_MixedPortionSizes tests preservation with mixed portion sizes
// **Validates: Requirements 3.1, 3.2, 3.4**
//
// IMPORTANT: This test runs on UNFIXED code to observe baseline behavior
// EXPECTED OUTCOME: Test PASSES (confirms baseline behavior to preserve)
//
// This test verifies that non-cooking status updates work correctly even with
// mixed portion sizes (SD schools with small and large portions).
func TestProperty3_Preservation_MixedPortionSizes(t *testing.T) {
	db := setupKDSBugfixTestDB(t)
	defer cleanupKDSBugfixTestDB(db)

	// Create KDS service directly without Firebase (for testing)
	kdsService := &KDSService{
		db:                db,
		firebaseApp:       nil,
		dbClient:          nil,
		monitoringService: nil,
	}

	properties := gopter.NewProperties(nil)

	// Property 3: Mixed portion sizes should not affect non-cooking status updates
	properties.Property("Mixed portion sizes preserve existing behavior for non-cooking updates", prop.ForAll(
		func(smallPortions, largePortions int) bool {
			cleanupKDSBugfixTestDB(db)
			ctx := context.Background()

			// Create test user
			user := models.User{
				NIK:          "1234567890",
				Email:        "chef@test.com",
				PasswordHash: "hashed_password",
				FullName:     "Test Chef",
				Role:         "chef",
				IsActive:     true,
			}
			if err := db.Create(&user).Error; err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create SD school (supports mixed portion sizes)
			school := models.School{
				Name:                "SD Test School",
				Address:             "Test Address",
				Latitude:            -6.2,
				Longitude:           106.8,
				StudentCountGrade13: 50,
				StudentCountGrade46: 50,
				Category:            "SD",
				IsActive:            true,
			}
			if err := db.Create(&school).Error; err != nil {
				t.Logf("Failed to create school: %v", err)
				return false
			}

			// Create semi-finished goods
			nasiGoods := models.SemiFinishedGoods{
				Name: "Nasi",
				Unit: "kg",
			}
			if err := db.Create(&nasiGoods).Error; err != nil {
				t.Logf("Failed to create nasi goods: %v", err)
				return false
			}

			// Create recipe
			recipe := models.Recipe{
				Name:          "Paket Nasi",
				Category:      "Paket",
				Instructions:  "Test recipe",
				TotalCalories: 500,
				TotalProtein:  20,
				TotalCarbs:    60,
				TotalFat:      15,
				CreatedBy:     user.ID,
			}
			if err := db.Create(&recipe).Error; err != nil {
				t.Logf("Failed to create recipe: %v", err)
				return false
			}

			// Create recipe item with portion-specific quantities
			nasiItem := models.RecipeItem{
				RecipeID:                recipe.ID,
				SemiFinishedGoodsID:     nasiGoods.ID,
				Quantity:                0,
				QuantityPerPortionSmall: 0.05,
				QuantityPerPortionLarge: 0.10,
			}
			if err := db.Create(&nasiItem).Error; err != nil {
				t.Logf("Failed to create nasi item: %v", err)
				return false
			}

			// Create inventory
			nasiInventory := models.SemiFinishedInventory{
				SemiFinishedGoodsID: nasiGoods.ID,
				Quantity:            20.0,
				LastUpdated:         time.Now(),
			}
			if err := db.Create(&nasiInventory).Error; err != nil {
				t.Logf("Failed to create nasi inventory: %v", err)
				return false
			}

			initialStock := nasiInventory.Quantity

			// Create menu plan
			today := time.Now()
			weekStart := today.AddDate(0, 0, -int(today.Weekday()))
			weekEnd := weekStart.AddDate(0, 0, 6)
			menuPlan := models.MenuPlan{
				WeekStart: weekStart,
				WeekEnd:   weekEnd,
				Status:    "approved",
				CreatedBy: user.ID,
			}
			if err := db.Create(&menuPlan).Error; err != nil {
				t.Logf("Failed to create menu plan: %v", err)
				return false
			}

			// Create menu item
			totalPortions := smallPortions + largePortions
			menuItem := models.MenuItem{
				MenuPlanID: menuPlan.ID,
				Date:       today,
				RecipeID:   recipe.ID,
				Portions:   totalPortions,
			}
			if err := db.Create(&menuItem).Error; err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Create school allocations with mixed portion sizes
			smallAllocation := models.MenuItemSchoolAllocation{
				MenuItemID:  menuItem.ID,
				SchoolID:    school.ID,
				Portions:    smallPortions,
				PortionSize: "small",
				Date:        today,
			}
			largeAllocation := models.MenuItemSchoolAllocation{
				MenuItemID:  menuItem.ID,
				SchoolID:    school.ID,
				Portions:    largePortions,
				PortionSize: "large",
				Date:        today,
			}
			if err := db.Create(&smallAllocation).Error; err != nil {
				t.Logf("Failed to create small allocation: %v", err)
				return false
			}
			if err := db.Create(&largeAllocation).Error; err != nil {
				t.Logf("Failed to create large allocation: %v", err)
				return false
			}

			// Test status update to "ready" (non-cooking)
			err := kdsService.UpdateRecipeStatus(ctx, recipe.ID, "ready", user.ID)

			// EXPECTED BEHAVIOR: Status update should succeed
			if err != nil {
				t.Logf("Unexpected error for 'ready' status with mixed portions: %v", err)
				return false
			}

			// Verify stock was NOT deducted
			var currentInventory models.SemiFinishedInventory
			db.Where("semi_finished_goods_id = ?", nasiGoods.ID).First(&currentInventory)

			if currentInventory.Quantity != initialStock {
				t.Logf("PRESERVATION VIOLATION: Stock changed for 'ready' status with mixed portions")
				t.Logf("  Initial: %.2f kg, Current: %.2f kg", initialStock, currentInventory.Quantity)
				t.Logf("  Small portions: %d, Large portions: %d", smallPortions, largePortions)
				return false
			}

			return true
		},
		gen.IntRange(10, 50), // smallPortions: 10-50
		gen.IntRange(10, 50), // largePortions: 10-50
	))

	properties.TestingRun(t)
}

// TestProperty4_Preservation_StatusSequence tests preservation across status sequences
// **Validates: Requirements 3.1, 3.2, 3.4**
//
// IMPORTANT: This test runs on UNFIXED code to observe baseline behavior
// EXPECTED OUTCOME: Test PASSES (confirms baseline behavior to preserve)
//
// This test verifies that sequences of non-cooking status updates work correctly
// and stock is never affected by these transitions.
func TestProperty4_Preservation_StatusSequence(t *testing.T) {
	db := setupKDSBugfixTestDB(t)
	defer cleanupKDSBugfixTestDB(db)

	// Create KDS service directly without Firebase (for testing)
	kdsService := &KDSService{
		db:                db,
		firebaseApp:       nil,
		dbClient:          nil,
		monitoringService: nil,
	}

	properties := gopter.NewProperties(nil)

	// Property 4: Sequences of non-cooking status updates should not affect stock
	properties.Property("Status sequences preserve stock levels", prop.ForAll(
		func(portions int) bool {
			cleanupKDSBugfixTestDB(db)
			ctx := context.Background()

			// Create test user
			user := models.User{
				NIK:          "1234567890",
				Email:        "chef@test.com",
				PasswordHash: "hashed_password",
				FullName:     "Test Chef",
				Role:         "chef",
				IsActive:     true,
			}
			if err := db.Create(&user).Error; err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create school
			school := models.School{
				Name:         "Test School",
				Address:      "Test Address",
				Latitude:     -6.2,
				Longitude:    106.8,
				StudentCount: 100,
				Category:     "SMP",
				IsActive:     true,
			}
			if err := db.Create(&school).Error; err != nil {
				t.Logf("Failed to create school: %v", err)
				return false
			}

			// Create semi-finished goods
			nasiGoods := models.SemiFinishedGoods{
				Name: "Nasi",
				Unit: "kg",
			}
			if err := db.Create(&nasiGoods).Error; err != nil {
				t.Logf("Failed to create nasi goods: %v", err)
				return false
			}

			// Create recipe
			recipe := models.Recipe{
				Name:          "Paket Nasi",
				Category:      "Paket",
				Instructions:  "Test recipe",
				TotalCalories: 500,
				TotalProtein:  20,
				TotalCarbs:    60,
				TotalFat:      15,
				CreatedBy:     user.ID,
			}
			if err := db.Create(&recipe).Error; err != nil {
				t.Logf("Failed to create recipe: %v", err)
				return false
			}

			// Create recipe item
			nasiItem := models.RecipeItem{
				RecipeID:                recipe.ID,
				SemiFinishedGoodsID:     nasiGoods.ID,
				Quantity:                0,
				QuantityPerPortionSmall: 0.05,
				QuantityPerPortionLarge: 0.10,
			}
			if err := db.Create(&nasiItem).Error; err != nil {
				t.Logf("Failed to create nasi item: %v", err)
				return false
			}

			// Create inventory
			nasiInventory := models.SemiFinishedInventory{
				SemiFinishedGoodsID: nasiGoods.ID,
				Quantity:            15.0,
				LastUpdated:         time.Now(),
			}
			if err := db.Create(&nasiInventory).Error; err != nil {
				t.Logf("Failed to create nasi inventory: %v", err)
				return false
			}

			initialStock := nasiInventory.Quantity

			// Create menu plan
			today := time.Now()
			weekStart := today.AddDate(0, 0, -int(today.Weekday()))
			weekEnd := weekStart.AddDate(0, 0, 6)
			menuPlan := models.MenuPlan{
				WeekStart: weekStart,
				WeekEnd:   weekEnd,
				Status:    "approved",
				CreatedBy: user.ID,
			}
			if err := db.Create(&menuPlan).Error; err != nil {
				t.Logf("Failed to create menu plan: %v", err)
				return false
			}

			// Create menu item
			menuItem := models.MenuItem{
				MenuPlanID: menuPlan.ID,
				Date:       today,
				RecipeID:   recipe.ID,
				Portions:   portions,
			}
			if err := db.Create(&menuItem).Error; err != nil {
				t.Logf("Failed to create menu item: %v", err)
				return false
			}

			// Create school allocation
			allocation := models.MenuItemSchoolAllocation{
				MenuItemID:  menuItem.ID,
				SchoolID:    school.ID,
				Portions:    portions,
				PortionSize: "large",
				Date:        today,
			}
			if err := db.Create(&allocation).Error; err != nil {
				t.Logf("Failed to create allocation: %v", err)
				return false
			}

			// Test sequence of status updates: pending -> ready -> pending
			statusSequence := []string{"pending", "ready", "pending"}
			for _, status := range statusSequence {
				err := kdsService.UpdateRecipeStatus(ctx, recipe.ID, status, user.ID)
				if err != nil {
					t.Logf("Unexpected error for status '%s' in sequence: %v", status, err)
					return false
				}

				// Verify stock remains unchanged after each transition
				var currentInventory models.SemiFinishedInventory
				db.Where("semi_finished_goods_id = ?", nasiGoods.ID).First(&currentInventory)

				if currentInventory.Quantity != initialStock {
					t.Logf("PRESERVATION VIOLATION: Stock changed during status sequence")
					t.Logf("  Status: %s, Initial: %.2f kg, Current: %.2f kg",
						status, initialStock, currentInventory.Quantity)
					return false
				}
			}

			// Verify no inventory movements were created throughout the sequence
			var movementCount int64
			db.Model(&models.InventoryMovement{}).Count(&movementCount)
			if movementCount > 0 {
				t.Logf("PRESERVATION VIOLATION: Inventory movements created during status sequence")
				return false
			}

			return true
		},
		gen.IntRange(20, 80), // portions: 20-80
	))

	properties.TestingRun(t)
}
