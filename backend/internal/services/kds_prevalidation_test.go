package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestPreValidationCheck tests that the pre-validation check collects all insufficient items
func TestPreValidationCheck(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Recipe{},
		&models.RecipeItem{},
		&models.SemiFinishedGoods{},
		&models.SemiFinishedInventory{},
		&models.InventoryMovement{},
		&models.MenuItem{},
		&models.MenuItemSchoolAllocation{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	// Create KDS service
	kdsService := &KDSService{
		db:                db,
		firebaseApp:       nil,
		dbClient:          nil,
		monitoringService: nil,
	}

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
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create semi-finished goods
	nasiGoods := models.SemiFinishedGoods{Name: "Nasi", Unit: "kg"}
	ayamGoods := models.SemiFinishedGoods{Name: "Ayam Goreng", Unit: "kg"}
	sayurGoods := models.SemiFinishedGoods{Name: "Sayur", Unit: "kg"}
	
	db.Create(&nasiGoods)
	db.Create(&ayamGoods)
	db.Create(&sayurGoods)

	// Create recipe
	recipe := models.Recipe{
		Name:          "Paket Lengkap",
		Category:      "Paket",
		Instructions:  "Test recipe",
		TotalCalories: 500,
		TotalProtein:  20,
		TotalCarbs:    60,
		TotalFat:      15,
		CreatedBy:     user.ID,
	}
	db.Create(&recipe)

	// Create recipe items
	nasiItem := models.RecipeItem{
		RecipeID:                recipe.ID,
		SemiFinishedGoodsID:     nasiGoods.ID,
		QuantityPerPortionSmall: 0.05,
		QuantityPerPortionLarge: 0.10,
	}
	ayamItem := models.RecipeItem{
		RecipeID:                recipe.ID,
		SemiFinishedGoodsID:     ayamGoods.ID,
		QuantityPerPortionSmall: 0.08,
		QuantityPerPortionLarge: 0.15,
	}
	sayurItem := models.RecipeItem{
		RecipeID:                recipe.ID,
		SemiFinishedGoodsID:     sayurGoods.ID,
		QuantityPerPortionSmall: 0.03,
		QuantityPerPortionLarge: 0.05,
	}
	db.Create(&nasiItem)
	db.Create(&ayamItem)
	db.Create(&sayurItem)

	// Create inventory with insufficient stock for multiple items
	// Need: 100 portions * 0.10 = 10 kg Nasi
	// Need: 100 portions * 0.15 = 15 kg Ayam
	// Need: 100 portions * 0.05 = 5 kg Sayur
	nasiInventory := models.SemiFinishedInventory{
		SemiFinishedGoodsID: nasiGoods.ID,
		Quantity:            5.0, // Insufficient (need 10)
		LastUpdated:         time.Now(),
	}
	ayamInventory := models.SemiFinishedInventory{
		SemiFinishedGoodsID: ayamGoods.ID,
		Quantity:            8.0, // Insufficient (need 15)
		LastUpdated:         time.Now(),
	}
	sayurInventory := models.SemiFinishedInventory{
		SemiFinishedGoodsID: sayurGoods.ID,
		Quantity:            10.0, // Sufficient (need 5)
		LastUpdated:         time.Now(),
	}
	db.Create(&nasiInventory)
	db.Create(&ayamInventory)
	db.Create(&sayurInventory)

	// Create menu item with school allocations
	menuItem := models.MenuItem{
		RecipeID: recipe.ID,
		Portions: 100,
	}
	db.Create(&menuItem)

	// Add school allocation (100 large portions)
	allocation := models.MenuItemSchoolAllocation{
		MenuItemID:  menuItem.ID,
		Portions:    100,
		PortionSize: "large",
	}
	db.Create(&allocation)
	menuItem.SchoolAllocations = append(menuItem.SchoolAllocations, allocation)

	// Load recipe with items
	db.Preload("RecipeItems").First(&menuItem.Recipe, recipe.ID)

	// Call deductInventory - should fail with detailed error
	ctx := context.Background()
	err = kdsService.deductInventory(ctx, &menuItem, user.ID)

	// Verify error is returned
	if err == nil {
		t.Fatal("Expected error for insufficient stock, but got nil")
	}

	// Verify error message contains all insufficient items
	errorMsg := err.Error()
	
	if !strings.Contains(errorMsg, "Stok tidak mencukupi untuk:") {
		t.Errorf("Error message should start with 'Stok tidak mencukupi untuk:', got: %s", errorMsg)
	}

	if !strings.Contains(errorMsg, "Nasi") {
		t.Errorf("Error message should mention Nasi, got: %s", errorMsg)
	}

	if !strings.Contains(errorMsg, "Ayam Goreng") {
		t.Errorf("Error message should mention Ayam Goreng, got: %s", errorMsg)
	}

	if strings.Contains(errorMsg, "Sayur") {
		t.Errorf("Error message should NOT mention Sayur (sufficient stock), got: %s", errorMsg)
	}

	// Verify error message contains needed and available quantities
	if !strings.Contains(errorMsg, "butuh") || !strings.Contains(errorMsg, "tersedia") {
		t.Errorf("Error message should contain 'butuh' and 'tersedia', got: %s", errorMsg)
	}

	// Verify stock was NOT deducted (transaction should not have started)
	var currentNasiInventory models.SemiFinishedInventory
	var currentAyamInventory models.SemiFinishedInventory
	var currentSayurInventory models.SemiFinishedInventory
	
	db.Where("semi_finished_goods_id = ?", nasiGoods.ID).First(&currentNasiInventory)
	db.Where("semi_finished_goods_id = ?", ayamGoods.ID).First(&currentAyamInventory)
	db.Where("semi_finished_goods_id = ?", sayurGoods.ID).First(&currentSayurInventory)

	if currentNasiInventory.Quantity != 5.0 {
		t.Errorf("Nasi stock should not be deducted, expected 5.0, got %.2f", currentNasiInventory.Quantity)
	}
	if currentAyamInventory.Quantity != 8.0 {
		t.Errorf("Ayam stock should not be deducted, expected 8.0, got %.2f", currentAyamInventory.Quantity)
	}
	if currentSayurInventory.Quantity != 10.0 {
		t.Errorf("Sayur stock should not be deducted, expected 10.0, got %.2f", currentSayurInventory.Quantity)
	}

	t.Logf("Pre-validation check working correctly. Error message: %s", errorMsg)
}
