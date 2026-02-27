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
		SchoolAllocations: []services.PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 200},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 100},
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
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 200},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 150},
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
		SchoolAllocations: []services.PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 200},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 100},
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
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 200},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 100}, // Sum = 300, but total = 350
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
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 200},
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
		SchoolAllocations: []services.PortionSizeAllocationInput{
			{SchoolID: school1.ID, PortionsSmall: 0, PortionsLarge: 200},
			{SchoolID: school2.ID, PortionsSmall: 0, PortionsLarge: 150},
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
	assert.Equal(t, float64(200), alloc1["total_portions"].(float64))
	assert.Equal(t, float64(0), alloc1["portions_small"].(float64))
	assert.Equal(t, float64(200), alloc1["portions_large"].(float64))

	alloc2 := allocations[1].(map[string]interface{})
	assert.Equal(t, "SD Negeri 2", alloc2["school_name"].(string))
	assert.Equal(t, float64(150), alloc2["total_portions"].(float64))
	assert.Equal(t, float64(0), alloc2["portions_small"].(float64))
	assert.Equal(t, float64(150), alloc2["portions_large"].(float64))
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
		SchoolAllocations: []services.PortionSizeAllocationInput{
			{SchoolID: school.ID, PortionsSmall: 0, PortionsLarge: 200},
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

// TestCreateMenuItem_ValidRequest_SDSchool tests successful creation with SD school (mixed portions)
func TestCreateMenuItem_ValidRequest_SDSchool(t *testing.T) {
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
	db.Create(&sdSchool)

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

	// Prepare create request with mixed portions for SD school
	createReq := CreateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 150},
		},
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(300), data["portions"])
	assert.Equal(t, float64(recipe.ID), data["recipe_id"])

	// Verify allocations were created
	allocations := data["school_allocations"].([]interface{})
	assert.Len(t, allocations, 2) // Should have 2 records: one small, one large

	// Verify database records
	var dbAllocations []models.MenuItemSchoolAllocation
	db.Where("school_id = ?", sdSchool.ID).Find(&dbAllocations)
	assert.Len(t, dbAllocations, 2)

	// Verify portion sizes
	smallFound := false
	largeFound := false
	for _, alloc := range dbAllocations {
		if alloc.PortionSize == "small" {
			assert.Equal(t, 150, alloc.Portions)
			smallFound = true
		} else if alloc.PortionSize == "large" {
			assert.Equal(t, 150, alloc.Portions)
			largeFound = true
		}
	}
	assert.True(t, smallFound, "Small portion allocation should exist")
	assert.True(t, largeFound, "Large portion allocation should exist")
}

// TestCreateMenuItem_ValidRequest_SMPSchool tests successful creation with SMP school (large only)
func TestCreateMenuItem_ValidRequest_SMPSchool(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	smpSchool := models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 200,
	}
	db.Create(&smpSchool)

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

	// Prepare create request with large portions only for SMP school
	createReq := CreateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
		},
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(200), data["portions"])

	// Verify allocations were created
	allocations := data["school_allocations"].([]interface{})
	assert.Len(t, allocations, 1) // Should have 1 record: large only

	// Verify database records
	var dbAllocations []models.MenuItemSchoolAllocation
	db.Where("school_id = ?", smpSchool.ID).Find(&dbAllocations)
	assert.Len(t, dbAllocations, 1)
	assert.Equal(t, "large", dbAllocations[0].PortionSize)
	assert.Equal(t, 200, dbAllocations[0].Portions)
}

// TestCreateMenuItem_ValidRequest_MultipleSchools tests creation with multiple schools
func TestCreateMenuItem_ValidRequest_MultipleSchools(t *testing.T) {
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

	// Prepare create request with multiple schools
	createReq := CreateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 500,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 150},
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
		},
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(500), data["portions"])

	// Verify allocations were created
	allocations := data["school_allocations"].([]interface{})
	assert.Len(t, allocations, 3) // Should have 3 records: 2 for SD (small+large), 1 for SMP (large)

	// Verify database records
	var dbAllocations []models.MenuItemSchoolAllocation
	db.Find(&dbAllocations)
	assert.Len(t, dbAllocations, 3)
}

// TestCreateMenuItem_InvalidSum tests rejection when sum doesn't match total
func TestCreateMenuItem_InvalidSum(t *testing.T) {
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
	db.Create(&sdSchool)

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

	// Prepare create request with invalid sum
	createReq := CreateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 300,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 100, PortionsLarge: 150}, // Sum = 250, but total = 300
		},
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "VALIDATION_ERROR", response["error_code"])
}

// TestCreateMenuItem_SMPWithSmallPortions tests rejection when SMP school has small portions
func TestCreateMenuItem_SMPWithSmallPortions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data
	smpSchool := models.School{
		Name:         "SMP Negeri 1",
		Category:     "SMP",
		StudentCount: 200,
	}
	db.Create(&smpSchool)

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

	// Prepare create request with small portions for SMP school (invalid)
	createReq := CreateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: smpSchool.ID, PortionsSmall: 50, PortionsLarge: 150}, // SMP cannot have small portions
		},
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "VALIDATION_ERROR", response["error_code"])
}

// TestCreateMenuItem_EmptyAllocations tests rejection when no allocations provided
func TestCreateMenuItem_EmptyAllocations(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

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

	// Prepare create request with empty allocations
	createReq := CreateMenuItemRequest{
		Date:              time.Now().Format("2006-01-02"),
		RecipeID:          recipe.ID,
		Portions:          200,
		SchoolAllocations: []SchoolAllocationInput{}, // Empty allocations
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response - should fail validation
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
}

// TestCreateMenuItem_NegativePortions tests rejection when portions are negative
func TestCreateMenuItem_NegativePortions(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	sdSchool := models.School{
		Name:                 "SD Negeri 1",
		Category:             "SD",
		StudentCount:         300,
		StudentCountGrade13:  150,
		StudentCountGrade46:  150,
	}
	db.Create(&sdSchool)

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

	// Prepare create request with negative portions
	createReq := CreateMenuItemRequest{
		Date:     time.Now().Format("2006-01-02"),
		RecipeID: recipe.ID,
		Portions: 200,
		SchoolAllocations: []SchoolAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: -50, PortionsLarge: 250}, // Negative small portions
		},
	}
	reqBody, _ := json.Marshal(createReq)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/menu-plans/%d/items", menuPlan.ID), bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
	}

	// Execute handler
	handler.CreateMenuItem(c)

	// Verify response - should fail validation
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
}

// TestGetMenuItem_WithPortionSizes tests GetMenuItem returns portion size breakdown
func TestGetMenuItem_WithPortionSizes(t *testing.T) {
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
		Portions: 500,
		SchoolAllocations: []services.PortionSizeAllocationInput{
			{SchoolID: sdSchool.ID, PortionsSmall: 150, PortionsLarge: 150},
			{SchoolID: smpSchool.ID, PortionsSmall: 0, PortionsLarge: 200},
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

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(500), data["portions"])

	// Verify school allocations with portion size breakdown
	allocations := data["school_allocations"].([]interface{})
	assert.Len(t, allocations, 2) // Should have 2 schools

	// Find SD school allocation
	var sdAlloc map[string]interface{}
	var smpAlloc map[string]interface{}
	for _, alloc := range allocations {
		allocMap := alloc.(map[string]interface{})
		if allocMap["school_category"].(string) == "SD" {
			sdAlloc = allocMap
		} else if allocMap["school_category"].(string) == "SMP" {
			smpAlloc = allocMap
		}
	}

	// Verify SD school has both small and large portions
	assert.NotNil(t, sdAlloc)
	assert.Equal(t, "SD Negeri 1", sdAlloc["school_name"])
	assert.Equal(t, "mixed", sdAlloc["portion_size_type"])
	assert.Equal(t, float64(150), sdAlloc["portions_small"])
	assert.Equal(t, float64(150), sdAlloc["portions_large"])
	assert.Equal(t, float64(300), sdAlloc["total_portions"])

	// Verify SMP school has only large portions
	assert.NotNil(t, smpAlloc)
	assert.Equal(t, "SMP Negeri 1", smpAlloc["school_name"])
	assert.Equal(t, "large", smpAlloc["portion_size_type"])
	assert.Equal(t, float64(0), smpAlloc["portions_small"])
	assert.Equal(t, float64(200), smpAlloc["portions_large"])
	assert.Equal(t, float64(200), smpAlloc["total_portions"])
}

// TestGetMenuItem_AlphabeticalOrder tests allocations are returned in alphabetical order
func TestGetMenuItem_AlphabeticalOrder(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	// Create test data - schools in non-alphabetical order
	schoolC := models.School{Name: "SD Negeri C", Category: "SD", StudentCount: 100}
	schoolA := models.School{Name: "SD Negeri A", Category: "SD", StudentCount: 100}
	schoolB := models.School{Name: "SD Negeri B", Category: "SD", StudentCount: 100}
	db.Create(&schoolC)
	db.Create(&schoolA)
	db.Create(&schoolB)

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
		Portions: 300,
		SchoolAllocations: []services.PortionSizeAllocationInput{
			{SchoolID: schoolC.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: schoolA.ID, PortionsSmall: 0, PortionsLarge: 100},
			{SchoolID: schoolB.ID, PortionsSmall: 0, PortionsLarge: 100},
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

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	allocations := data["school_allocations"].([]interface{})
	assert.Len(t, allocations, 3)

	// Verify alphabetical order
	names := []string{}
	for _, alloc := range allocations {
		allocMap := alloc.(map[string]interface{})
		names = append(names, allocMap["school_name"].(string))
	}
	assert.Equal(t, "SD Negeri A", names[0])
	assert.Equal(t, "SD Negeri B", names[1])
	assert.Equal(t, "SD Negeri C", names[2])
}

// TestGetMenuItem_NotFound_WithPortionSizes tests 404 response when menu item doesn't exist
func TestGetMenuItem_NotFound_WithPortionSizes(t *testing.T) {
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

	// Create test request with non-existent item ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/menu-plans/%d/items/99999", menuPlan.ID), nil)
	c.Params = gin.Params{
		{Key: "id", Value: fmt.Sprintf("%d", menuPlan.ID)},
		{Key: "item_id", Value: "99999"},
	}

	// Execute handler
	handler.GetMenuItem(c)

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "NOT_FOUND", response["error_code"])
}
