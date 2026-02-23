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

type ProductionWorkflowTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	token  string
}

func (suite *ProductionWorkflowTestSuite) SetupSuite() {
	// Setup test database
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test_production",
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

func (suite *ProductionWorkflowTestSuite) TearDownSuite() {
	// Clean up test database
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *ProductionWorkflowTestSuite) SetupTest() {
	// Clean up data before each test
	suite.db.Exec("DELETE FROM delivery_menu_items")
	suite.db.Exec("DELETE FROM delivery_tasks")
	suite.db.Exec("DELETE FROM electronic_pods")
	suite.db.Exec("DELETE FROM menu_items")
	suite.db.Exec("DELETE FROM menu_plans")
	suite.db.Exec("DELETE FROM recipe_ingredients")
	suite.db.Exec("DELETE FROM recipes")
	suite.db.Exec("DELETE FROM ingredients")
	suite.db.Exec("DELETE FROM schools")
	suite.db.Exec("DELETE FROM inventory_movements")
	suite.db.Exec("DELETE FROM inventory_items")
}

func (suite *ProductionWorkflowTestSuite) setupTestUser() {
	// Create test user
	user := &models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$test.hash",
		FullName:     "Test User",
		Role:         "ahli_gizi",
		IsActive:     true,
	}
	suite.db.Create(user)
	
	// Login to get token
	loginData := map[string]string{
		"email":    "test@example.com",
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

func (suite *ProductionWorkflowTestSuite) makeAuthenticatedRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
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

func (suite *ProductionWorkflowTestSuite) TestCompleteProductionWorkflow() {
	// Test complete flow: Menu planning → Cooking → Packing → Delivery
	
	// Step 1: Setup master data
	suite.setupMasterData()
	
	// Step 2: Create menu plan
	menuPlanID := suite.createMenuPlan()
	
	// Step 3: Approve menu plan
	suite.approveMenuPlan(menuPlanID)
	
	// Step 4: Start cooking process
	recipeID := suite.startCookingProcess()
	
	// Step 5: Complete cooking and move to packing
	suite.completeCooking(recipeID)
	
	// Step 6: Complete packing
	schoolID := suite.completePacking()
	
	// Step 7: Create delivery task
	deliveryTaskID := suite.createDeliveryTask(schoolID)
	
	// Step 8: Complete delivery
	suite.completeDelivery(deliveryTaskID)
	
	// Step 9: Verify data consistency across all modules
	suite.verifyDataConsistency(menuPlanID, recipeID, schoolID, deliveryTaskID)
}

func (suite *ProductionWorkflowTestSuite) setupMasterData() {
	// Create ingredient
	ingredient := &models.Ingredient{
		Name:            "Beras",
		Unit:            "kg",
		CaloriesPer100g: 130,
		ProteinPer100g:  2.7,
		CarbsPer100g:    28,
		FatPer100g:      0.3,
	}
	suite.db.Create(ingredient)
	
	// Create recipe
	recipe := &models.Recipe{
		Name:          "Nasi Putih",
		Category:      "Makanan Pokok",
		ServingSize:   100,
		Instructions:  "Masak nasi hingga matang",
		TotalCalories: 130,
		TotalProtein:  2.7,
		TotalCarbs:    28,
		TotalFat:      0.3,
		Version:       1,
		IsActive:      true,
		CreatedBy:     1,
	}
	suite.db.Create(recipe)
	
	// Create recipe ingredient
	recipeIngredient := &models.RecipeIngredient{
		RecipeID:     recipe.ID,
		IngredientID: ingredient.ID,
		Quantity:     1.0,
	}
	suite.db.Create(recipeIngredient)
	
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
	
	// Create inventory
	inventory := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     1000,
		MinThreshold: 100,
		LastUpdated:  time.Now(),
	}
	suite.db.Create(inventory)
}

func (suite *ProductionWorkflowTestSuite) createMenuPlan() uint {
	menuPlanData := map[string]interface{}{
		"week_start": time.Now().Format("2006-01-02"),
		"week_end":   time.Now().AddDate(0, 0, 6).Format("2006-01-02"),
		"menu_items": []map[string]interface{}{
			{
				"date":      time.Now().Format("2006-01-02"),
				"recipe_id": 1,
				"portions":  200,
			},
		},
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/menu-plans", menuPlanData)
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	return uint(response["id"].(float64))
}

func (suite *ProductionWorkflowTestSuite) approveMenuPlan(menuPlanID uint) {
	url := fmt.Sprintf("/api/v1/menu-plans/%d/approve", menuPlanID)
	w := suite.makeAuthenticatedRequest("POST", url, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	// Verify menu plan is approved
	var menuPlan models.MenuPlan
	suite.db.First(&menuPlan, menuPlanID)
	suite.Equal("approved", menuPlan.Status)
	suite.NotNil(menuPlan.ApprovedAt)
}

func (suite *ProductionWorkflowTestSuite) startCookingProcess() uint {
	// Get today's menu from KDS
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/kds/cooking/today", nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	recipes := response["recipes"].([]interface{})
	suite.Greater(len(recipes), 0)
	
	recipe := recipes[0].(map[string]interface{})
	recipeID := uint(recipe["id"].(float64))
	
	// Start cooking
	url := fmt.Sprintf("/api/v1/kds/cooking/%d/status", recipeID)
	statusData := map[string]string{"status": "cooking"}
	
	w = suite.makeAuthenticatedRequest("PUT", url, statusData)
	suite.Equal(http.StatusOK, w.Code)
	
	return recipeID
}

func (suite *ProductionWorkflowTestSuite) completeCooking(recipeID uint) {
	// Complete cooking
	url := fmt.Sprintf("/api/v1/kds/cooking/%d/status", recipeID)
	statusData := map[string]string{"status": "ready"}
	
	w := suite.makeAuthenticatedRequest("PUT", url, statusData)
	suite.Equal(http.StatusOK, w.Code)
	
	// Verify inventory was deducted
	var inventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", 1).First(&inventory)
	suite.Less(inventory.Quantity, float64(1000)) // Should be less than initial amount
}

func (suite *ProductionWorkflowTestSuite) completePacking() uint {
	// Get packing allocations
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/kds/packing/today", nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	allocations := response["allocations"].([]interface{})
	suite.Greater(len(allocations), 0)
	
	allocation := allocations[0].(map[string]interface{})
	schoolID := uint(allocation["school_id"].(float64))
	
	// Complete packing
	url := fmt.Sprintf("/api/v1/kds/packing/%d/status", schoolID)
	statusData := map[string]string{"status": "ready"}
	
	w = suite.makeAuthenticatedRequest("PUT", url, statusData)
	suite.Equal(http.StatusOK, w.Code)
	
	return schoolID
}

func (suite *ProductionWorkflowTestSuite) createDeliveryTask(schoolID uint) uint {
	deliveryTaskData := map[string]interface{}{
		"task_date":  time.Now().Format("2006-01-02"),
		"driver_id":  1,
		"school_id":  schoolID,
		"portions":   200,
		"menu_items": []map[string]interface{}{
			{
				"recipe_id": 1,
				"portions":  200,
			},
		},
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/delivery-tasks", deliveryTaskData)
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	return uint(response["id"].(float64))
}

func (suite *ProductionWorkflowTestSuite) completeDelivery(deliveryTaskID uint) {
	// Create e-POD
	epodData := map[string]interface{}{
		"delivery_task_id": deliveryTaskID,
		"photo_url":        "https://example.com/photo.jpg",
		"signature_url":    "https://example.com/signature.jpg",
		"latitude":         -6.2088,
		"longitude":        106.8456,
		"recipient_name":   "Kepala Sekolah",
		"ompreng_drop_off": 10,
		"ompreng_pick_up":  8,
		"completed_at":     time.Now().Format(time.RFC3339),
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/epod", epodData)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Verify delivery task status updated
	var deliveryTask models.DeliveryTask
	suite.db.First(&deliveryTask, deliveryTaskID)
	suite.Equal("completed", deliveryTask.Status)
}

func (suite *ProductionWorkflowTestSuite) verifyDataConsistency(menuPlanID, recipeID, schoolID, deliveryTaskID uint) {
	// Verify menu plan exists and is approved
	var menuPlan models.MenuPlan
	suite.db.Preload("MenuItems").First(&menuPlan, menuPlanID)
	suite.Equal("approved", menuPlan.Status)
	suite.Greater(len(menuPlan.MenuItems), 0)
	
	// Verify recipe exists and was used
	var recipe models.Recipe
	suite.db.First(&recipe, recipeID)
	suite.True(recipe.IsActive)
	
	// Verify school exists
	var school models.School
	suite.db.First(&school, schoolID)
	suite.True(school.IsActive)
	
	// Verify delivery task completed
	var deliveryTask models.DeliveryTask
	suite.db.Preload("MenuItems").First(&deliveryTask, deliveryTaskID)
	suite.Equal("completed", deliveryTask.Status)
	suite.Greater(len(deliveryTask.MenuItems), 0)
	
	// Verify e-POD created
	var epod models.ElectronicPOD
	suite.db.Where("delivery_task_id = ?", deliveryTaskID).First(&epod)
	suite.NotEmpty(epod.PhotoURL)
	suite.NotEmpty(epod.SignatureURL)
	
	// Verify inventory movement recorded
	var movements []models.InventoryMovement
	suite.db.Where("ingredient_id = ? AND movement_type = ?", 1, "out").Find(&movements)
	suite.Greater(len(movements), 0)
	
	// Verify ompreng tracking
	var omprengTracking models.OmprengTracking
	suite.db.Where("school_id = ?", schoolID).First(&omprengTracking)
	suite.Equal(10, omprengTracking.DropOff)
	suite.Equal(8, omprengTracking.PickUp)
	suite.Equal(2, omprengTracking.Balance) // 10 - 8 = 2
}

func (suite *ProductionWorkflowTestSuite) TestWorkflowDataIntegrity() {
	// Test that data remains consistent throughout the workflow
	suite.setupMasterData()
	
	// Create menu plan
	menuPlanID := suite.createMenuPlan()
	
	// Verify initial state
	var initialInventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", 1).First(&initialInventory)
	initialQuantity := initialInventory.Quantity
	
	// Complete workflow
	suite.approveMenuPlan(menuPlanID)
	recipeID := suite.startCookingProcess()
	suite.completeCooking(recipeID)
	
	// Verify inventory deduction
	var finalInventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", 1).First(&finalInventory)
	
	// Should have deducted exactly the amount needed for the recipe
	expectedDeduction := float64(200) // 200 portions * 1kg per 100 portions = 2kg
	suite.InDelta(initialQuantity-expectedDeduction/100, finalInventory.Quantity, 0.1)
}

func (suite *ProductionWorkflowTestSuite) TestWorkflowErrorHandling() {
	// Test workflow behavior when errors occur
	suite.setupMasterData()
	
	// Try to start cooking without approved menu
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/kds/cooking/today", nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	recipes := response["recipes"].([]interface{})
	suite.Equal(0, len(recipes)) // Should be empty without approved menu
	
	// Try to complete delivery without e-POD
	deliveryTaskData := map[string]interface{}{
		"task_date": time.Now().Format("2006-01-02"),
		"driver_id": 1,
		"school_id": 1,
		"portions":  200,
	}
	
	w = suite.makeAuthenticatedRequest("POST", "/api/v1/delivery-tasks", deliveryTaskData)
	suite.Equal(http.StatusCreated, w.Code)
	
	var taskResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &taskResponse)
	taskID := uint(taskResponse["id"].(float64))
	
	// Verify task is not completed without e-POD
	var deliveryTask models.DeliveryTask
	suite.db.First(&deliveryTask, taskID)
	suite.NotEqual("completed", deliveryTask.Status)
}

func TestProductionWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(ProductionWorkflowTestSuite))
}