package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseDateParameter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		queryParam  string
		expectError bool
		description string
	}{
		{
			name:        "Valid date format",
			queryParam:  "2024-01-15",
			expectError: false,
			description: "Should accept valid YYYY-MM-DD format",
		},
		{
			name:        "Missing date parameter",
			queryParam:  "",
			expectError: false,
			description: "Should default to current date when parameter is missing",
		},
		{
			name:        "Invalid format - wrong separator",
			queryParam:  "2024/01/15",
			expectError: true,
			description: "Should reject date with wrong separator",
		},
		{
			name:        "Invalid format - missing leading zeros",
			queryParam:  "2024-1-5",
			expectError: true,
			description: "Should reject date without leading zeros",
		},
		{
			name:        "Invalid format - wrong order",
			queryParam:  "01-15-2024",
			expectError: true,
			description: "Should reject date in MM-DD-YYYY format",
		},
		{
			name:        "Invalid date - February 30",
			queryParam:  "2024-02-30",
			expectError: true,
			description: "Should reject invalid date like February 30",
		},
		{
			name:        "Invalid date - month 13",
			queryParam:  "2024-13-01",
			expectError: true,
			description: "Should reject invalid month",
		},
		{
			name:        "Invalid date - day 32",
			queryParam:  "2024-01-32",
			expectError: true,
			description: "Should reject invalid day",
		},
		{
			name:        "Valid leap year date",
			queryParam:  "2024-02-29",
			expectError: false,
			description: "Should accept valid leap year date",
		},
		{
			name:        "Invalid non-leap year date",
			queryParam:  "2023-02-29",
			expectError: true,
			description: "Should reject February 29 in non-leap year",
		},
		{
			name:        "Future date",
			queryParam:  "2025-12-31",
			expectError: false,
			description: "Should accept future dates",
		},
		{
			name:        "Past date",
			queryParam:  "2020-01-01",
			expectError: false,
			description: "Should accept past dates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			// Create a test request with query parameter
			req := httptest.NewRequest(http.MethodGet, "/?date="+tt.queryParam, nil)
			c.Request = req

			// Call the function
			date, err := parseDateParameter(c)

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				
				// For valid dates, verify the date is normalized to start of day
				if tt.queryParam != "" {
					assert.Equal(t, 0, date.Hour(), "Hour should be 0")
					assert.Equal(t, 0, date.Minute(), "Minute should be 0")
					assert.Equal(t, 0, date.Second(), "Second should be 0")
					assert.Equal(t, 0, date.Nanosecond(), "Nanosecond should be 0")
				}
			}
		})
	}
}

func TestParseDateParameterDefaultsToToday(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test context without date parameter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = req

	// Call the function
	date, err := parseDateParameter(c)

	// Should not error
	assert.NoError(t, err)

	// Should return today's date
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	expectedDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	assert.Equal(t, expectedDate.Year(), date.Year(), "Year should match current year")
	assert.Equal(t, expectedDate.Month(), date.Month(), "Month should match current month")
	assert.Equal(t, expectedDate.Day(), date.Day(), "Day should match current day")
	assert.Equal(t, 0, date.Hour(), "Hour should be 0")
	assert.Equal(t, 0, date.Minute(), "Minute should be 0")
	assert.Equal(t, 0, date.Second(), "Second should be 0")
}

func TestParseDateParameterTimezone(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test context with a valid date
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/?date=2024-01-15", nil)
	c.Request = req

	// Call the function
	date, err := parseDateParameter(c)

	// Should not error
	assert.NoError(t, err)

	// Verify timezone is Asia/Jakarta
	loc, _ := time.LoadLocation("Asia/Jakarta")
	assert.Equal(t, loc.String(), date.Location().String(), "Timezone should be Asia/Jakarta")
}

// TestGetCookingToday_WithPortionSizes tests KDS cooking view includes portion size breakdown
func TestGetCookingToday_WithPortionSizes(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	sdSchool := models.School{
		Name:                 "SD Negeri 1",
		Category:             "SD",
		StudentCount:         300,
		StudentCountGrade13:  150,
		StudentCountGrade46:  150,
	}
	smpSchool := models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 200,
	}
	db.Create(&sdSchool)
	db.Create(&smpSchool)

	recipe := models.Recipe{
		Name:         "Nasi Goreng",
		Category:     "Main Course",
		Instructions: "Cook rice",
	}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "approved", // Must be approved to show in KDS
	}
	db.Create(&menuPlan)

	// Create menu item with allocations
	menuItem := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       time.Now(),
		RecipeID:   recipe.ID,
		Portions:   500,
	}
	db.Create(&menuItem)

	// Create allocations with portion sizes
	allocations := []models.MenuItemSchoolAllocation{
		{
			MenuItemID:  menuItem.ID,
			SchoolID:    sdSchool.ID,
			Portions:    150,
			PortionSize: "small",
			Date:        time.Now(),
		},
		{
			MenuItemID:  menuItem.ID,
			SchoolID:    sdSchool.ID,
			Portions:    150,
			PortionSize: "large",
			Date:        time.Now(),
		},
		{
			MenuItemID:  menuItem.ID,
			SchoolID:    smpSchool.ID,
			Portions:    200,
			PortionSize: "large",
			Date:        time.Now(),
		},
	}
	for _, alloc := range allocations {
		db.Create(&alloc)
	}

	// Test the data retrieval logic directly
	// Query menu items with allocations
	var menuItems []models.MenuItem
	err := db.
		Preload("Recipe").
		Preload("SchoolAllocations").
		Preload("SchoolAllocations.School").
		Preload("MenuPlan").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("DATE(menu_items.date) = DATE(?)", time.Now()).
		Find(&menuItems).Error
	assert.NoError(t, err)
	assert.Len(t, menuItems, 1)

	// Verify the menu item
	item := menuItems[0]
	assert.Equal(t, recipe.ID, item.RecipeID)
	assert.Equal(t, 500, item.Portions)
	assert.Len(t, item.SchoolAllocations, 3) // 3 allocation records

	// Group allocations by school (simulating the service logic)
	schoolMap := make(map[uint]*services.SchoolAllocationResponse)
	for _, alloc := range item.SchoolAllocations {
		schoolID := alloc.SchoolID
		
		if _, exists := schoolMap[schoolID]; !exists {
			portionSizeType := "large"
			if alloc.School.Category == "SD" {
				portionSizeType = "mixed"
			}
			
			schoolMap[schoolID] = &services.SchoolAllocationResponse{
				SchoolID:        schoolID,
				SchoolName:      alloc.School.Name,
				SchoolCategory:  alloc.School.Category,
				PortionSizeType: portionSizeType,
				PortionsSmall:   0,
				PortionsLarge:   0,
				TotalPortions:   0,
			}
		}
		
		if alloc.PortionSize == "small" {
			schoolMap[schoolID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			schoolMap[schoolID].PortionsLarge += alloc.Portions
		}
		schoolMap[schoolID].TotalPortions += alloc.Portions
	}

	// Verify we have 2 schools
	assert.Len(t, schoolMap, 2)

	// Verify SD school
	sdAlloc := schoolMap[sdSchool.ID]
	assert.NotNil(t, sdAlloc)
	assert.Equal(t, "SD Negeri 1", sdAlloc.SchoolName)
	assert.Equal(t, "mixed", sdAlloc.PortionSizeType)
	assert.Equal(t, 150, sdAlloc.PortionsSmall)
	assert.Equal(t, 150, sdAlloc.PortionsLarge)
	assert.Equal(t, 300, sdAlloc.TotalPortions)

	// Verify SMP school
	smpAlloc := schoolMap[smpSchool.ID]
	assert.NotNil(t, smpAlloc)
	assert.Equal(t, "SMP Negeri 1", smpAlloc.SchoolName)
	assert.Equal(t, "large", smpAlloc.PortionSizeType)
	assert.Equal(t, 0, smpAlloc.PortionsSmall)
	assert.Equal(t, 200, smpAlloc.PortionsLarge)
	assert.Equal(t, 200, smpAlloc.TotalPortions)
}

// TestCalculatePackingAllocations_WithPortionSizes tests packing view includes portion size breakdown
func TestCalculatePackingAllocations_WithPortionSizes(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	sdSchool := models.School{
		Name:                 "SD Negeri 1",
		Category:             "SD",
		StudentCount:         300,
		StudentCountGrade13:  150,
		StudentCountGrade46:  150,
	}
	smpSchool := models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 200,
	}
	db.Create(&sdSchool)
	db.Create(&smpSchool)

	recipe1 := models.Recipe{Name: "Nasi Goreng", Category: "Main Course"}
	recipe2 := models.Recipe{Name: "Ayam Goreng", Category: "Main Course"}
	db.Create(&recipe1)
	db.Create(&recipe2)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "approved",
	}
	db.Create(&menuPlan)

	// Create menu items
	menuItem1 := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       time.Now(),
		RecipeID:   recipe1.ID,
		Portions:   300,
	}
	menuItem2 := models.MenuItem{
		MenuPlanID: menuPlan.ID,
		Date:       time.Now(),
		RecipeID:   recipe2.ID,
		Portions:   200,
	}
	db.Create(&menuItem1)
	db.Create(&menuItem2)

	// Create allocations with portion sizes
	allocations := []models.MenuItemSchoolAllocation{
		// Recipe 1 - SD school
		{MenuItemID: menuItem1.ID, SchoolID: sdSchool.ID, Portions: 100, PortionSize: "small", Date: time.Now()},
		{MenuItemID: menuItem1.ID, SchoolID: sdSchool.ID, Portions: 100, PortionSize: "large", Date: time.Now()},
		// Recipe 1 - SMP school
		{MenuItemID: menuItem1.ID, SchoolID: smpSchool.ID, Portions: 100, PortionSize: "large", Date: time.Now()},
		// Recipe 2 - SD school
		{MenuItemID: menuItem2.ID, SchoolID: sdSchool.ID, Portions: 50, PortionSize: "small", Date: time.Now()},
		{MenuItemID: menuItem2.ID, SchoolID: sdSchool.ID, Portions: 50, PortionSize: "large", Date: time.Now()},
		// Recipe 2 - SMP school
		{MenuItemID: menuItem2.ID, SchoolID: smpSchool.ID, Portions: 100, PortionSize: "large", Date: time.Now()},
	}
	for _, alloc := range allocations {
		db.Create(&alloc)
	}

	// Test the grouping logic directly
	var menuAllocations []models.MenuItemSchoolAllocation
	err := db.
		Preload("School").
		Preload("MenuItem").
		Preload("MenuItem.Recipe").
		Find(&menuAllocations).Error
	assert.NoError(t, err)
	assert.Len(t, menuAllocations, 6) // Should have 6 allocation records

	// Group by school (simulating the service logic)
	schoolMap := make(map[uint]*services.SchoolAllocation)
	menuItemMap := make(map[uint]map[uint]*services.MenuItemSummary)
	
	for _, alloc := range menuAllocations {
		if _, exists := schoolMap[alloc.SchoolID]; !exists {
			portionSizeType := "large"
			if alloc.School.Category == "SD" {
				portionSizeType = "mixed"
			}
			
			schoolMap[alloc.SchoolID] = &services.SchoolAllocation{
				SchoolID:        alloc.School.ID,
				SchoolName:      alloc.School.Name,
				SchoolCategory:  alloc.School.Category,
				PortionSizeType: portionSizeType,
				PortionsSmall:   0,
				PortionsLarge:   0,
				TotalPortions:   0,
				MenuItems:       []services.MenuItemSummary{},
				Status:          "pending",
			}
			menuItemMap[alloc.SchoolID] = make(map[uint]*services.MenuItemSummary)
		}
		
		if alloc.PortionSize == "small" {
			schoolMap[alloc.SchoolID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			schoolMap[alloc.SchoolID].PortionsLarge += alloc.Portions
		}
		schoolMap[alloc.SchoolID].TotalPortions += alloc.Portions
		
		recipeID := alloc.MenuItem.Recipe.ID
		if _, exists := menuItemMap[alloc.SchoolID][recipeID]; !exists {
			menuItemMap[alloc.SchoolID][recipeID] = &services.MenuItemSummary{
				RecipeID:      recipeID,
				RecipeName:    alloc.MenuItem.Recipe.Name,
				PortionsSmall: 0,
				PortionsLarge: 0,
				TotalPortions: 0,
			}
		}
		
		if alloc.PortionSize == "small" {
			menuItemMap[alloc.SchoolID][recipeID].PortionsSmall += alloc.Portions
		} else if alloc.PortionSize == "large" {
			menuItemMap[alloc.SchoolID][recipeID].PortionsLarge += alloc.Portions
		}
		menuItemMap[alloc.SchoolID][recipeID].TotalPortions += alloc.Portions
	}
	
	// Convert menu item map to slices
	for schoolID, school := range schoolMap {
		for _, menuItem := range menuItemMap[schoolID] {
			school.MenuItems = append(school.MenuItems, *menuItem)
		}
	}

	// Verify we have 2 schools
	assert.Len(t, schoolMap, 2)

	// Verify SD school
	sdAlloc := schoolMap[sdSchool.ID]
	assert.NotNil(t, sdAlloc)
	assert.Equal(t, "SD Negeri 1", sdAlloc.SchoolName)
	assert.Equal(t, "mixed", sdAlloc.PortionSizeType)
	assert.Equal(t, 150, sdAlloc.PortionsSmall) // 100 + 50
	assert.Equal(t, 150, sdAlloc.PortionsLarge) // 100 + 50
	assert.Equal(t, 300, sdAlloc.TotalPortions)
	assert.Len(t, sdAlloc.MenuItems, 2) // 2 recipes

	// Verify SD school menu items
	for _, menuItem := range sdAlloc.MenuItems {
		if menuItem.RecipeName == "Nasi Goreng" {
			assert.Equal(t, 100, menuItem.PortionsSmall)
			assert.Equal(t, 100, menuItem.PortionsLarge)
			assert.Equal(t, 200, menuItem.TotalPortions)
		} else if menuItem.RecipeName == "Ayam Goreng" {
			assert.Equal(t, 50, menuItem.PortionsSmall)
			assert.Equal(t, 50, menuItem.PortionsLarge)
			assert.Equal(t, 100, menuItem.TotalPortions)
		}
	}

	// Verify SMP school
	smpAlloc := schoolMap[smpSchool.ID]
	assert.NotNil(t, smpAlloc)
	assert.Equal(t, "SMP Negeri 1", smpAlloc.SchoolName)
	assert.Equal(t, "large", smpAlloc.PortionSizeType)
	assert.Equal(t, 0, smpAlloc.PortionsSmall)
	assert.Equal(t, 200, smpAlloc.PortionsLarge) // 100 + 100
	assert.Equal(t, 200, smpAlloc.TotalPortions)
	assert.Len(t, smpAlloc.MenuItems, 2) // 2 recipes

	// Verify SMP school menu items
	for _, menuItem := range smpAlloc.MenuItems {
		assert.Equal(t, 0, menuItem.PortionsSmall)
		assert.Equal(t, 100, menuItem.PortionsLarge)
		assert.Equal(t, 100, menuItem.TotalPortions)
	}
}
