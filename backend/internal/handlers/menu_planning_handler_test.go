package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.School{},
		&models.Recipe{},
		&models.MenuPlan{},
		&models.MenuItem{},
		&models.MenuItemSchoolAllocation{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TestUpdateMenuItem_ValidUpdate tests successful update with valid allocations
func TestUpdateMenuItem_ValidUpdate(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	school1 := models.School{Name: "SD Negeri 1", StudentCount: 200}
	school2 := models.School{Name: "SD Negeri 2", StudentCount: 150}
	db.Create(&school1)
	db.Create(&school2)

	recipe := models.Recipe{Name: "Nasi Goreng", Category: "Main Course"}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
	}
	db.Create(&menuPlan)

	// Create initial menu item
	service := services.NewMenuPlanningService(db)
	initialInput := services.MenuItemInput{
		Date:     time.Now(),
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []services.SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 200},
			{SchoolID: school2.ID, Portions: 100},
		},
	}
	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	assert.NoError(t, err)

	// Create handler
	handler := NewMenuPlanningHandler(db)

	// Prepare update request
	updateReq := UpdateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 350,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 200},
			{SchoolID: school2.ID, Portions: 150},
		},
	}
	reqBody, _ := json.Marshal(updateReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/menu-plans/%d/items/%d", menuPlan.ID, menuItem.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
		{Key: "item_id", Value: fmt.Sprintf("%d", menuItem.ID)},
	}

	// Execute handler
	handler.UpdateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(350), data["portions"])

	allocations := data["school_allocations"].([]interface{})
	assert.Len(t, allocations, 2)
}

// TestUpdateMenuItem_InvalidSum tests rejection when sum doesn't match total
func TestUpdateMenuItem_InvalidSum(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	school1 := models.School{Name: "SD Negeri 1", StudentCount: 200}
	school2 := models.School{Name: "SD Negeri 2", StudentCount: 150}
	db.Create(&school1)
	db.Create(&school2)

	recipe := models.Recipe{Name: "Nasi Goreng", Category: "Main Course"}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
	}
	db.Create(&menuPlan)

	// Create initial menu item
	service := services.NewMenuPlanningService(db)
	initialInput := services.MenuItemInput{
		Date:     time.Now(),
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []services.SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 200},
			{SchoolID: school2.ID, Portions: 100},
		},
	}
	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, initialInput)
	assert.NoError(t, err)

	// Create handler
	handler := NewMenuPlanningHandler(db)

	// Prepare update request with invalid sum
	updateReq := UpdateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 350,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 200},
			{SchoolID: school2.ID, Portions: 100}, // Sum = 300, but total = 350
		},
	}
	reqBody, _ := json.Marshal(updateReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/menu-plans/%d/items/%d", menuPlan.ID, menuItem.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
		{Key: "item_id", Value: fmt.Sprintf("%d", menuItem.ID)},
	}

	// Execute handler
	handler.UpdateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "VALIDATION_ERROR", response["error_code"])
}

// TestUpdateMenuItem_NonExistentMenuItem tests rejection when menu item doesn't exist
func TestUpdateMenuItem_NonExistentMenuItem(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	school1 := models.School{Name: "SD Negeri 1", StudentCount: 200}
	db.Create(&school1)

	recipe := models.Recipe{Name: "Nasi Goreng", Category: "Main Course"}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
	}
	db.Create(&menuPlan)

	// Create handler
	handler := NewMenuPlanningHandler(db)

	// Prepare update request for non-existent menu item
	updateReq := UpdateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 200},
		},
	}
	reqBody, _ := json.Marshal(updateReq)

	// Create test request with non-existent item_id
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/menu-plans/%d/items/999", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
		{Key: "item_id", Value: "999"},
	}

	// Execute handler
	handler.UpdateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "VALIDATION_ERROR", response["error_code"])
}

// TestGetMenuItem_Success tests successful retrieval of menu item with allocations
func TestGetMenuItem_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	school1 := models.School{Name: "SD Negeri 1", StudentCount: 200}
	school2 := models.School{Name: "SD Negeri 2", StudentCount: 150}
	db.Create(&school1)
	db.Create(&school2)

	recipe := models.Recipe{Name: "Nasi Goreng", Category: "Main Course"}
	db.Create(&recipe)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
	}
	db.Create(&menuPlan)

	// Create menu item with allocations
	service := services.NewMenuPlanningService(db)
	input := services.MenuItemInput{
		Date:     time.Now(),
		RecipeID: recipe.ID,
		Portions: 350,
		SchoolAllocations: []services.SchoolAllocationInput{
			{SchoolID: school1.ID, Portions: 200},
			{SchoolID: school2.ID, Portions: 150},
		},
	}
	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan.ID, input)
	assert.NoError(t, err)

	// Create handler
	handler := NewMenuPlanningHandler(db)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/menu-plans/%d/items/%d", menuPlan.ID, menuItem.ID), nil)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
		{Key: "item_id", Value: fmt.Sprintf("%d", menuItem.ID)},
	}

	// Execute handler
	handler.GetMenuItem(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	// Verify data structure
	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(menuItem.ID), data["id"].(float64))
	assert.Equal(t, float64(menuPlan.ID), data["menu_plan_id"].(float64))
	assert.Equal(t, float64(recipe.ID), data["recipe_id"].(float64))
	assert.Equal(t, float64(350), data["portions"].(float64))

	// Verify recipe data
	recipeData := data["recipe"].(map[string]interface{})
	assert.Equal(t, "Nasi Goreng", recipeData["name"].(string))

	// Verify school allocations
	allocations := data["school_allocations"].([]interface{})
	assert.Equal(t, 2, len(allocations))

	// Verify allocations are ordered by school name (SD Negeri 1, SD Negeri 2)
	alloc1 := allocations[0].(map[string]interface{})
	assert.Equal(t, "SD Negeri 1", alloc1["school_name"].(string))
	assert.Equal(t, float64(200), alloc1["portions"].(float64))

	alloc2 := allocations[1].(map[string]interface{})
	assert.Equal(t, "SD Negeri 2", alloc2["school_name"].(string))
	assert.Equal(t, float64(150), alloc2["portions"].(float64))
}

// TestGetMenuItem_NotFound tests retrieval of non-existent menu item
func TestGetMenuItem_NotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	menuPlan := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
	}
	db.Create(&menuPlan)

	// Create handler
	handler := NewMenuPlanningHandler(db)

	// Create test request with non-existent item_id
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/menu-plans/%d/items/999", menuPlan.ID), nil)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
		{Key: "item_id", Value: "999"},
	}

	// Execute handler
	handler.GetMenuItem(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "NOT_FOUND", response["error_code"].(string))
}

// TestGetMenuItem_WrongMenuPlan tests retrieval of menu item from wrong menu plan
func TestGetMenuItem_WrongMenuPlan(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	school := models.School{Name: "SD Negeri 1", StudentCount: 200}
	db.Create(&school)

	recipe := models.Recipe{Name: "Nasi Goreng", Category: "Main Course"}
	db.Create(&recipe)

	menuPlan1 := models.MenuPlan{
		WeekStart: time.Now(),
		WeekEnd:   time.Now().AddDate(0, 0, 7),
		Status:    "draft",
	}
	db.Create(&menuPlan1)

	menuPlan2 := models.MenuPlan{
		WeekStart: time.Now().AddDate(0, 0, 7),
		WeekEnd:   time.Now().AddDate(0, 0, 14),
		Status:    "draft",
	}
	db.Create(&menuPlan2)

	// Create menu item in menuPlan1
	service := services.NewMenuPlanningService(db)
	input := services.MenuItemInput{
		Date:     time.Now(),
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []services.SchoolAllocationInput{
			{SchoolID: school.ID, Portions: 200},
		},
	}
	menuItem, err := service.CreateMenuItemWithAllocations(menuPlan1.ID, input)
	assert.NoError(t, err)

	// Create handler
	handler := NewMenuPlanningHandler(db)

	// Try to get menu item using menuPlan2 ID (wrong menu plan)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/menu-plans/%d/items/%d", menuPlan2.ID, menuItem.ID), nil)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan2.ID)},
		{Key: "item_id", Value: fmt.Sprintf("%d", menuItem.ID)},
	}

	// Execute handler
	handler.GetMenuItem(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "NOT_FOUND", response["error_code"].(string))
}
