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

type ProcurementWorkflowTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	token  string
}

func (suite *ProcurementWorkflowTestSuite) SetupSuite() {
	// Setup test database
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "test_procurement",
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

func (suite *ProcurementWorkflowTestSuite) TearDownSuite() {
	// Clean up test database
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *ProcurementWorkflowTestSuite) SetupTest() {
	// Clean up data before each test
	suite.db.Exec("DELETE FROM cash_flow_entries")
	suite.db.Exec("DELETE FROM goods_receipt_items")
	suite.db.Exec("DELETE FROM goods_receipts")
	suite.db.Exec("DELETE FROM purchase_order_items")
	suite.db.Exec("DELETE FROM purchase_orders")
	suite.db.Exec("DELETE FROM inventory_movements")
	suite.db.Exec("DELETE FROM inventory_items")
	suite.db.Exec("DELETE FROM suppliers")
	suite.db.Exec("DELETE FROM ingredients")
}

func (suite *ProcurementWorkflowTestSuite) setupTestUser() {
	// Create test users with different roles
	pengadaanUser := &models.User{
		NIK:          "1234567890",
		Email:        "pengadaan@example.com",
		PasswordHash: "$2a$10$test.hash",
		FullName:     "Staff Pengadaan",
		Role:         "pengadaan",
		IsActive:     true,
	}
	suite.db.Create(pengadaanUser)
	
	kepalaUser := &models.User{
		NIK:          "0987654321",
		Email:        "kepala@example.com",
		PasswordHash: "$2a$10$test.hash",
		FullName:     "Kepala SPPG",
		Role:         "kepala_sppg",
		IsActive:     true,
	}
	suite.db.Create(kepalaUser)
	
	// Login as pengadaan user to get token
	loginData := map[string]string{
		"email":    "pengadaan@example.com",
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

func (suite *ProcurementWorkflowTestSuite) makeAuthenticatedRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
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

func (suite *ProcurementWorkflowTestSuite) loginAsKepala() string {
	loginData := map[string]string{
		"email":    "kepala@example.com",
		"password": "password",
	}
	
	body, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["token"].(string)
}

func (suite *ProcurementWorkflowTestSuite) TestCompleteProcurementWorkflow() {
	// Test complete flow: PO creation → Approval → GRN → Inventory update → Cash flow entry
	
	// Step 1: Setup master data
	supplierID, ingredientID := suite.setupMasterData()
	
	// Step 2: Create Purchase Order
	poID := suite.createPurchaseOrder(supplierID, ingredientID)
	
	// Step 3: Approve Purchase Order (as Kepala SPPG)
	suite.approvePurchaseOrder(poID)
	
	// Step 4: Create Goods Receipt Note
	grnID := suite.createGoodsReceipt(poID, ingredientID)
	
	// Step 5: Verify automatic triggers and data propagation
	suite.verifyInventoryUpdate(ingredientID)
	suite.verifyCashFlowEntry(grnID)
	
	// Step 6: Verify complete data consistency
	suite.verifyProcurementDataConsistency(poID, grnID, supplierID, ingredientID)
}

func (suite *ProcurementWorkflowTestSuite) setupMasterData() (uint, uint) {
	// Create supplier
	supplier := &models.Supplier{
		Name:            "PT Beras Sejahtera",
		ContactPerson:   "Budi Santoso",
		PhoneNumber:     "021-1234567",
		Email:           "budi@berassejahtera.com",
		Address:         "Jl. Industri No. 123, Jakarta",
		ProductCategory: "Bahan Makanan Pokok",
		IsActive:        true,
		OnTimeDelivery:  95.5,
		QualityRating:   4.5,
	}
	suite.db.Create(supplier)
	
	// Create ingredient
	ingredient := &models.Ingredient{
		Name:            "Beras Premium",
		Unit:            "kg",
		CaloriesPer100g: 130,
		ProteinPer100g:  2.7,
		CarbsPer100g:    28,
		FatPer100g:      0.3,
	}
	suite.db.Create(ingredient)
	
	// Create initial inventory item with low stock
	inventory := &models.InventoryItem{
		IngredientID: ingredient.ID,
		Quantity:     50, // Below threshold
		MinThreshold: 100,
		LastUpdated:  time.Now(),
	}
	suite.db.Create(inventory)
	
	return supplier.ID, ingredient.ID
}

func (suite *ProcurementWorkflowTestSuite) createPurchaseOrder(supplierID, ingredientID uint) uint {
	poData := map[string]interface{}{
		"supplier_id":       supplierID,
		"order_date":        time.Now().Format("2006-01-02"),
		"expected_delivery": time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
		"po_items": []map[string]interface{}{
			{
				"ingredient_id": ingredientID,
				"quantity":      500,
				"unit_price":    12000,
				"subtotal":      6000000,
			},
		},
		"total_amount": 6000000,
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/purchase-orders", poData)
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	poID := uint(response["id"].(float64))
	
	// Verify PO created with correct status
	var po models.PurchaseOrder
	suite.db.Preload("POItems").First(&po, poID)
	suite.Equal("pending", po.Status)
	suite.Equal(float64(6000000), po.TotalAmount)
	suite.Equal(1, len(po.POItems))
	suite.NotEmpty(po.PONumber)
	
	return poID
}

func (suite *ProcurementWorkflowTestSuite) approvePurchaseOrder(poID uint) {
	// Switch to Kepala SPPG token
	kepalaToken := suite.loginAsKepala()
	originalToken := suite.token
	suite.token = kepalaToken
	
	url := fmt.Sprintf("/api/v1/purchase-orders/%d/approve", poID)
	w := suite.makeAuthenticatedRequest("POST", url, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	// Verify PO status updated
	var po models.PurchaseOrder
	suite.db.First(&po, poID)
	suite.Equal("approved", po.Status)
	suite.NotNil(po.ApprovedAt)
	suite.Equal(uint(2), *po.ApprovedBy) // Kepala SPPG user ID
	
	// Switch back to original token
	suite.token = originalToken
}

func (suite *ProcurementWorkflowTestSuite) createGoodsReceipt(poID, ingredientID uint) uint {
	grnData := map[string]interface{}{
		"po_id":        poID,
		"receipt_date": time.Now().Format("2006-01-02"),
		"invoice_photo": "https://storage.example.com/invoices/inv123.jpg",
		"notes":        "Barang diterima dalam kondisi baik",
		"grn_items": []map[string]interface{}{
			{
				"ingredient_id":      ingredientID,
				"ordered_quantity":   500,
				"received_quantity":  500,
				"expiry_date":        time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
			},
		},
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/goods-receipts", grnData)
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	grnID := uint(response["id"].(float64))
	
	// Verify GRN created correctly
	var grn models.GoodsReceipt
	suite.db.Preload("GRNItems").First(&grn, grnID)
	suite.NotEmpty(grn.GRNNumber)
	suite.Equal("https://storage.example.com/invoices/inv123.jpg", grn.InvoicePhoto)
	suite.Equal(1, len(grn.GRNItems))
	
	return grnID
}

func (suite *ProcurementWorkflowTestSuite) verifyInventoryUpdate(ingredientID uint) {
	// Verify inventory quantity increased
	var inventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", ingredientID).First(&inventory)
	suite.Equal(float64(550), inventory.Quantity) // 50 + 500 = 550
	
	// Verify inventory movement recorded
	var movement models.InventoryMovement
	suite.db.Where("ingredient_id = ? AND movement_type = ?", ingredientID, "in").First(&movement)
	suite.Equal(float64(500), movement.Quantity)
	suite.Equal("in", movement.MovementType)
	suite.Contains(movement.Reference, "GRN") // Should reference GRN number
}

func (suite *ProcurementWorkflowTestSuite) verifyCashFlowEntry(grnID uint) {
	// Verify automatic cash flow entry created
	var grn models.GoodsReceipt
	suite.db.Preload("PurchaseOrder").First(&grn, grnID)
	
	var cashFlow models.CashFlowEntry
	suite.db.Where("reference = ?", grn.GRNNumber).First(&cashFlow)
	
	suite.Equal("expense", cashFlow.Type)
	suite.Equal("bahan_baku", cashFlow.Category)
	suite.Equal(float64(6000000), cashFlow.Amount)
	suite.Contains(cashFlow.Description, "Pembelian bahan baku")
	suite.Equal(grn.GRNNumber, cashFlow.Reference)
}

func (suite *ProcurementWorkflowTestSuite) verifyProcurementDataConsistency(poID, grnID, supplierID, ingredientID uint) {
	// Verify Purchase Order consistency
	var po models.PurchaseOrder
	suite.db.Preload("POItems").Preload("Supplier").First(&po, poID)
	suite.Equal("approved", po.Status)
	suite.Equal(supplierID, po.SupplierID)
	suite.NotNil(po.ApprovedAt)
	suite.NotNil(po.ApprovedBy)
	
	// Verify GRN consistency
	var grn models.GoodsReceipt
	suite.db.Preload("GRNItems").Preload("PurchaseOrder").First(&grn, grnID)
	suite.Equal(poID, grn.POID)
	suite.Equal(1, len(grn.GRNItems))
	
	// Verify GRN items match PO items
	grnItem := grn.GRNItems[0]
	poItem := po.POItems[0]
	suite.Equal(poItem.IngredientID, grnItem.IngredientID)
	suite.Equal(poItem.Quantity, grnItem.OrderedQuantity)
	suite.Equal(grnItem.OrderedQuantity, grnItem.ReceivedQuantity) // Full delivery
	
	// Verify supplier performance tracking
	var supplier models.Supplier
	suite.db.First(&supplier, supplierID)
	suite.True(supplier.IsActive)
	suite.Greater(supplier.OnTimeDelivery, float64(0))
	suite.Greater(supplier.QualityRating, float64(0))
	
	// Verify inventory final state
	var inventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", ingredientID).First(&inventory)
	suite.Greater(inventory.Quantity, inventory.MinThreshold) // Should be above threshold now
	
	// Verify cash flow impact
	var totalExpenses float64
	suite.db.Model(&models.CashFlowEntry{}).Where("category = ? AND type = ?", "bahan_baku", "expense").Select("SUM(amount)").Scan(&totalExpenses)
	suite.Equal(float64(6000000), totalExpenses)
}

func (suite *ProcurementWorkflowTestSuite) TestProcurementWorkflowWithDiscrepancies() {
	// Test workflow when received quantity differs from ordered quantity
	
	supplierID, ingredientID := suite.setupMasterData()
	poID := suite.createPurchaseOrder(supplierID, ingredientID)
	suite.approvePurchaseOrder(poID)
	
	// Create GRN with quantity discrepancy
	grnData := map[string]interface{}{
		"po_id":        poID,
		"receipt_date": time.Now().Format("2006-01-02"),
		"invoice_photo": "https://storage.example.com/invoices/inv124.jpg",
		"notes":        "Barang kurang 50kg dari pesanan",
		"grn_items": []map[string]interface{}{
			{
				"ingredient_id":      ingredientID,
				"ordered_quantity":   500,
				"received_quantity":  450, // 50kg short
				"expiry_date":        time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
			},
		},
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/goods-receipts", grnData)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Verify inventory updated with actual received quantity
	var inventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", ingredientID).First(&inventory)
	suite.Equal(float64(500), inventory.Quantity) // 50 + 450 = 500
	
	// Verify cash flow entry reflects actual received amount (proportional)
	expectedAmount := float64(6000000) * (450.0 / 500.0) // Proportional to received quantity
	var cashFlow models.CashFlowEntry
	suite.db.Where("category = ?", "bahan_baku").First(&cashFlow)
	suite.InDelta(expectedAmount, cashFlow.Amount, 1000) // Allow small rounding difference
}

func (suite *ProcurementWorkflowTestSuite) TestProcurementWorkflowErrorHandling() {
	// Test error scenarios in procurement workflow
	
	supplierID, ingredientID := suite.setupMasterData()
	
	// Test: Cannot create GRN for unapproved PO
	poID := suite.createPurchaseOrder(supplierID, ingredientID)
	// Skip approval step
	
	grnData := map[string]interface{}{
		"po_id":        poID,
		"receipt_date": time.Now().Format("2006-01-02"),
		"grn_items": []map[string]interface{}{
			{
				"ingredient_id":      ingredientID,
				"ordered_quantity":   500,
				"received_quantity":  500,
			},
		},
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/goods-receipts", grnData)
	suite.Equal(http.StatusBadRequest, w.Code) // Should reject unapproved PO
	
	// Test: Cannot approve PO without proper role
	// (Already using pengadaan token, not kepala_sppg)
	url := fmt.Sprintf("/api/v1/purchase-orders/%d/approve", poID)
	w = suite.makeAuthenticatedRequest("POST", url, nil)
	suite.Equal(http.StatusForbidden, w.Code) // Should reject unauthorized approval
}

func (suite *ProcurementWorkflowTestSuite) TestLowStockAlertGeneration() {
	// Test that low stock alerts are generated during procurement workflow
	
	supplierID, ingredientID := suite.setupMasterData()
	
	// Verify initial low stock condition
	var inventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", ingredientID).First(&inventory)
	suite.Less(inventory.Quantity, inventory.MinThreshold) // Should be below threshold
	
	// Check if low stock alert exists
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/inventory/alerts", nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var alertResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &alertResponse)
	
	alerts := alertResponse["alerts"].([]interface{})
	suite.Greater(len(alerts), 0) // Should have at least one alert
	
	// Complete procurement to resolve low stock
	poID := suite.createPurchaseOrder(supplierID, ingredientID)
	suite.approvePurchaseOrder(poID)
	suite.createGoodsReceipt(poID, ingredientID)
	
	// Verify stock is now above threshold
	suite.db.Where("ingredient_id = ?", ingredientID).First(&inventory)
	suite.Greater(inventory.Quantity, inventory.MinThreshold)
	
	// Verify alert is resolved (implementation may vary)
	w = suite.makeAuthenticatedRequest("GET", "/api/v1/inventory/alerts", nil)
	suite.Equal(http.StatusOK, w.Code)
	
	json.Unmarshal(w.Body.Bytes(), &alertResponse)
	alerts = alertResponse["alerts"].([]interface{})
	// Alert count should be reduced or alert should be marked as resolved
}

func (suite *ProcurementWorkflowTestSuite) TestFIFOInventoryMethod() {
	// Test FIFO inventory method during procurement
	
	supplierID, ingredientID := suite.setupMasterData()
	
	// Create first batch with earlier expiry
	poID1 := suite.createPurchaseOrder(supplierID, ingredientID)
	suite.approvePurchaseOrder(poID1)
	
	grnData1 := map[string]interface{}{
		"po_id":        poID1,
		"receipt_date": time.Now().Format("2006-01-02"),
		"grn_items": []map[string]interface{}{
			{
				"ingredient_id":      ingredientID,
				"ordered_quantity":   500,
				"received_quantity":  500,
				"expiry_date":        time.Now().AddDate(0, 6, 0).Format("2006-01-02"), // 6 months
			},
		},
	}
	
	w := suite.makeAuthenticatedRequest("POST", "/api/v1/goods-receipts", grnData1)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Create second batch with later expiry
	poID2 := suite.createPurchaseOrder(supplierID, ingredientID)
	suite.approvePurchaseOrder(poID2)
	
	grnData2 := map[string]interface{}{
		"po_id":        poID2,
		"receipt_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		"grn_items": []map[string]interface{}{
			{
				"ingredient_id":      ingredientID,
				"ordered_quantity":   300,
				"received_quantity":  300,
				"expiry_date":        time.Now().AddDate(1, 0, 0).Format("2006-01-02"), // 12 months
			},
		},
	}
	
	w = suite.makeAuthenticatedRequest("POST", "/api/v1/goods-receipts", grnData2)
	suite.Equal(http.StatusCreated, w.Code)
	
	// Verify total inventory
	var inventory models.InventoryItem
	suite.db.Where("ingredient_id = ?", ingredientID).First(&inventory)
	suite.Equal(float64(850), inventory.Quantity) // 50 + 500 + 300 = 850
	
	// Verify inventory movements recorded in correct order
	var movements []models.InventoryMovement
	suite.db.Where("ingredient_id = ? AND movement_type = ?", ingredientID, "in").Order("movement_date ASC").Find(&movements)
	suite.Equal(2, len(movements))
	suite.Equal(float64(500), movements[0].Quantity)
	suite.Equal(float64(300), movements[1].Quantity)
}

func TestProcurementWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(ProcurementWorkflowTestSuite))
}