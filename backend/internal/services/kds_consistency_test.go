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

// TestKDSDataConsistency verifies that portion size data is consistent across KDS Cooking View and KDS Packing View
func TestKDSDataConsistency(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.School{},
		&models.Recipe{},
		&models.SemiFinishedGoods{},
		&models.RecipeItem{},
		&models.MenuPlan{},
		&models.MenuItem{},
		&models.MenuItemSchoolAllocation{},
	)
	require.NoError(t, err)

	// Create test data
	ctx := context.Background()
	testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create schools
	sdSchool := models.School{
		Name:                "SD Negeri 1",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Address:             "Test Address",
		ContactPerson:       "Test Contact",
		PhoneNumber:         "123456789",
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	err = db.Create(&sdSchool).Error
	require.NoError(t, err)

	smpSchool := models.School{
		Name:          "SMP Negeri 1",
		Category:      "SMP",
		StudentCount:  200,
		Address:       "Test Address",
		ContactPerson: "Test Contact",
		PhoneNumber:   "123456789",
		Latitude:      -6.2,
		Longitude:     106.8,
	}
	err = db.Create(&smpSchool).Error
	require.NoError(t, err)

	// Create recipe
	recipe := models.Recipe{
		Name:         "Nasi Goreng",
		Instructions: "Cook rice",
	}
	err = db.Create(&recipe).Error
	require.NoError(t, err)

	// Create menu plan
	menuPlan := models.MenuPlan{
		WeekStart: testDate.AddDate(0, 0, -7),
		WeekEnd:   testDate.AddDate(0, 0, 7),
		Status:    "approved",
		CreatedBy: 1,
	}
	err = db.Create(&menuPlan).Error
	require.NoError(t, err)

	// Create menu item
	menuItem := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		RecipeID:   recipe.ID,
		Date:       testDate,
		Portions:   500,
	}
	err = db.Create(&menuItem).Error
	require.NoError(t, err)

	// Create allocations with portion sizes
	allocations := []models.MenuItemSchoolAllocation{
		{
			MenuItemID:  menuItem.ID,
			SchoolID:    sdSchool.ID,
			Portions:    150,
			PortionSize: "small",
			Date:        testDate,
		},
		{
			MenuItemID:  menuItem.ID,
			SchoolID:    sdSchool.ID,
			Portions:    200,
			PortionSize: "large",
			Date:        testDate,
		},
		{
			MenuItemID:  menuItem.ID,
			SchoolID:    smpSchool.ID,
			Portions:    150,
			PortionSize: "large",
			Date:        testDate,
		},
	}
	for _, alloc := range allocations {
		err = db.Create(&alloc).Error
		require.NoError(t, err)
	}

	// Create KDS service (without Firebase for this test)
	kdsService := &KDSService{
		db: db,
	}

	// Create packing allocation service (without Firebase for this test)
	packingService := &PackingAllocationService{
		db: db,
	}

	// Test 1: Get cooking view data
	cookingData, err := kdsService.GetTodayMenu(ctx, testDate)
	require.NoError(t, err)
	require.Len(t, cookingData, 1, "Should have one recipe")

	recipe1 := cookingData[0]
	assert.Equal(t, "Nasi Goreng", recipe1.Name)
	assert.Equal(t, 500, recipe1.PortionsRequired)
	require.Len(t, recipe1.SchoolAllocations, 2, "Should have 2 schools")

	// Test 2: Get packing view data
	packingData, err := packingService.CalculatePackingAllocations(ctx, testDate)
	require.NoError(t, err)
	require.Len(t, packingData, 2, "Should have 2 schools")

	// Test 3: Verify SD school data consistency
	var sdCookingAlloc *SchoolAllocationResponse
	for _, alloc := range recipe1.SchoolAllocations {
		if alloc.SchoolID == sdSchool.ID {
			allocCopy := alloc
			sdCookingAlloc = &allocCopy
			break
		}
	}
	require.NotNil(t, sdCookingAlloc, "SD school should be in cooking view")

	var sdPackingAlloc *SchoolAllocation
	for _, alloc := range packingData {
		if alloc.SchoolID == sdSchool.ID {
			allocCopy := alloc
			sdPackingAlloc = &allocCopy
			break
		}
	}
	require.NotNil(t, sdPackingAlloc, "SD school should be in packing view")

	// Verify SD school portion sizes match
	assert.Equal(t, sdCookingAlloc.SchoolName, sdPackingAlloc.SchoolName, "School names should match")
	assert.Equal(t, sdCookingAlloc.SchoolCategory, sdPackingAlloc.SchoolCategory, "School categories should match")
	assert.Equal(t, sdCookingAlloc.PortionSizeType, sdPackingAlloc.PortionSizeType, "Portion size types should match")
	assert.Equal(t, sdCookingAlloc.PortionsSmall, sdPackingAlloc.PortionsSmall, "Small portions should match")
	assert.Equal(t, sdCookingAlloc.PortionsLarge, sdPackingAlloc.PortionsLarge, "Large portions should match")
	assert.Equal(t, sdCookingAlloc.TotalPortions, sdPackingAlloc.TotalPortions, "Total portions should match")

	// Verify SD school specific values
	assert.Equal(t, "mixed", sdCookingAlloc.PortionSizeType, "SD school should have mixed portion type")
	assert.Equal(t, 150, sdCookingAlloc.PortionsSmall, "SD school should have 150 small portions")
	assert.Equal(t, 200, sdCookingAlloc.PortionsLarge, "SD school should have 200 large portions")
	assert.Equal(t, 350, sdCookingAlloc.TotalPortions, "SD school should have 350 total portions")

	// Test 4: Verify SMP school data consistency
	var smpCookingAlloc *SchoolAllocationResponse
	for _, alloc := range recipe1.SchoolAllocations {
		if alloc.SchoolID == smpSchool.ID {
			allocCopy := alloc
			smpCookingAlloc = &allocCopy
			break
		}
	}
	require.NotNil(t, smpCookingAlloc, "SMP school should be in cooking view")

	var smpPackingAlloc *SchoolAllocation
	for _, alloc := range packingData {
		if alloc.SchoolID == smpSchool.ID {
			allocCopy := alloc
			smpPackingAlloc = &allocCopy
			break
		}
	}
	require.NotNil(t, smpPackingAlloc, "SMP school should be in packing view")

	// Verify SMP school portion sizes match
	assert.Equal(t, smpCookingAlloc.SchoolName, smpPackingAlloc.SchoolName, "School names should match")
	assert.Equal(t, smpCookingAlloc.SchoolCategory, smpPackingAlloc.SchoolCategory, "School categories should match")
	assert.Equal(t, smpCookingAlloc.PortionSizeType, smpPackingAlloc.PortionSizeType, "Portion size types should match")
	assert.Equal(t, smpCookingAlloc.PortionsSmall, smpPackingAlloc.PortionsSmall, "Small portions should match (should be 0)")
	assert.Equal(t, smpCookingAlloc.PortionsLarge, smpPackingAlloc.PortionsLarge, "Large portions should match")
	assert.Equal(t, smpCookingAlloc.TotalPortions, smpPackingAlloc.TotalPortions, "Total portions should match")

	// Verify SMP school specific values
	assert.Equal(t, "large", smpCookingAlloc.PortionSizeType, "SMP school should have large portion type")
	assert.Equal(t, 0, smpCookingAlloc.PortionsSmall, "SMP school should have 0 small portions")
	assert.Equal(t, 150, smpCookingAlloc.PortionsLarge, "SMP school should have 150 large portions")
	assert.Equal(t, 150, smpCookingAlloc.TotalPortions, "SMP school should have 150 total portions")

	// Test 5: Verify total portions consistency
	totalCookingPortions := 0
	for _, alloc := range recipe1.SchoolAllocations {
		totalCookingPortions += alloc.TotalPortions
	}

	totalPackingPortions := 0
	for _, alloc := range packingData {
		totalPackingPortions += alloc.TotalPortions
	}

	assert.Equal(t, totalCookingPortions, totalPackingPortions, "Total portions should match between cooking and packing views")
	assert.Equal(t, 500, totalCookingPortions, "Total portions should equal menu item portions")
}

// TestKDSDataConsistencyMultipleRecipes verifies consistency with multiple recipes
func TestKDSDataConsistencyMultipleRecipes(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.School{},
		&models.Recipe{},
		&models.SemiFinishedGoods{},
		&models.RecipeItem{},
		&models.MenuPlan{},
		&models.MenuItem{},
		&models.MenuItemSchoolAllocation{},
	)
	require.NoError(t, err)

	ctx := context.Background()
	testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create school
	sdSchool := models.School{
		Name:                "SD Negeri 1",
		Category:            "SD",
		StudentCount:        300,
		StudentCountGrade13: 150,
		StudentCountGrade46: 150,
		Address:             "Test Address",
		ContactPerson:       "Test Contact",
		PhoneNumber:         "123456789",
		Latitude:            -6.2,
		Longitude:           106.8,
	}
	err = db.Create(&sdSchool).Error
	require.NoError(t, err)

	// Create menu plan
	menuPlan := models.MenuPlan{
		WeekStart: testDate.AddDate(0, 0, -7),
		WeekEnd:   testDate.AddDate(0, 0, 7),
		Status:    "approved",
		CreatedBy: 1,
	}
	err = db.Create(&menuPlan).Error
	require.NoError(t, err)

	// Create multiple recipes and menu items
	recipes := []struct {
		name          string
		portionsSmall int
		portionsLarge int
		totalPortions int
	}{
		{"Nasi Goreng", 100, 150, 250},
		{"Mie Goreng", 50, 100, 150},
	}

	for _, r := range recipes {
		recipe := models.Recipe{
			Name:         r.name,
			Instructions: "Cook",
		}
		err = db.Create(&recipe).Error
		require.NoError(t, err)

		menuItem := models.MenuItem{
			MenuPlanID: menuPlan.ID,
			RecipeID:   recipe.ID,
			Date:       testDate,
			Portions:   r.totalPortions,
		}
		err = db.Create(&menuItem).Error
		require.NoError(t, err)

		// Create allocations
		allocations := []models.MenuItemSchoolAllocation{
			{
				MenuItemID:  menuItem.ID,
				SchoolID:    sdSchool.ID,
				Portions:    r.portionsSmall,
				PortionSize: "small",
				Date:        testDate,
			},
			{
				MenuItemID:  menuItem.ID,
				SchoolID:    sdSchool.ID,
				Portions:    r.portionsLarge,
				PortionSize: "large",
				Date:        testDate,
			},
		}
		for _, alloc := range allocations {
			err = db.Create(&alloc).Error
			require.NoError(t, err)
		}
	}

	// Create services
	kdsService := &KDSService{db: db}
	packingService := &PackingAllocationService{db: db}

	// Get data from both views
	cookingData, err := kdsService.GetTodayMenu(ctx, testDate)
	require.NoError(t, err)
	require.Len(t, cookingData, 2, "Should have 2 recipes")

	packingData, err := packingService.CalculatePackingAllocations(ctx, testDate)
	require.NoError(t, err)
	require.Len(t, packingData, 1, "Should have 1 school")

	// Verify packing view aggregates all recipes correctly
	sdPacking := packingData[0]
	assert.Equal(t, "SD Negeri 1", sdPacking.SchoolName)
	assert.Equal(t, "mixed", sdPacking.PortionSizeType)

	// Calculate expected totals from cooking view
	expectedSmall := 0
	expectedLarge := 0
	for _, recipe := range cookingData {
		for _, alloc := range recipe.SchoolAllocations {
			if alloc.SchoolID == sdSchool.ID {
				expectedSmall += alloc.PortionsSmall
				expectedLarge += alloc.PortionsLarge
			}
		}
	}

	// Verify packing view matches cooking view totals
	assert.Equal(t, expectedSmall, sdPacking.PortionsSmall, "Packing view should aggregate small portions from all recipes")
	assert.Equal(t, expectedLarge, sdPacking.PortionsLarge, "Packing view should aggregate large portions from all recipes")
	assert.Equal(t, expectedSmall+expectedLarge, sdPacking.TotalPortions, "Total portions should match sum")

	// Verify specific values
	assert.Equal(t, 150, sdPacking.PortionsSmall, "Should have 150 total small portions (100+50)")
	assert.Equal(t, 250, sdPacking.PortionsLarge, "Should have 250 total large portions (150+100)")
	assert.Equal(t, 400, sdPacking.TotalPortions, "Should have 400 total portions")
}
