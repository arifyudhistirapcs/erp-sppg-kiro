package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// uploadBaseDir is the base directory for file uploads
var uploadBaseDir = filepath.Join("uploads")

// SupplyChainHandler handles supply chain endpoints
type SupplyChainHandler struct {
	supplierService       *services.SupplierService
	purchaseOrderService  *services.PurchaseOrderService
	goodsReceiptService   *services.GoodsReceiptService
	inventoryService      *services.InventoryService
}

// NewSupplyChainHandler creates a new supply chain handler
func NewSupplyChainHandler(db *gorm.DB) *SupplyChainHandler {
	inventoryService := services.NewInventoryService(db)
	cashFlowService := services.NewCashFlowService(db)
	
	return &SupplyChainHandler{
		supplierService:      services.NewSupplierService(db),
		purchaseOrderService: services.NewPurchaseOrderService(db),
		goodsReceiptService:  services.NewGoodsReceiptService(db, inventoryService, cashFlowService),
		inventoryService:     inventoryService,
	}
}

// Supplier Endpoints

// CreateSupplierRequest represents create supplier request
type CreateSupplierRequest struct {
	Name            string  `json:"name" binding:"required"`
	ContactPerson   string  `json:"contact_person"`
	PhoneNumber     string  `json:"phone_number"`
	Email           string  `json:"email" binding:"omitempty,email"`
	Address         string  `json:"address"`
	ProductCategory string  `json:"product_category"`
}

// CreateSupplier creates a new supplier
func (h *SupplyChainHandler) CreateSupplier(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	supplier := &models.Supplier{
		Name:            req.Name,
		ContactPerson:   req.ContactPerson,
		PhoneNumber:     req.PhoneNumber,
		Email:           req.Email,
		Address:         req.Address,
		ProductCategory: req.ProductCategory,
	}

	if err := h.supplierService.CreateSupplier(supplier); err != nil {
		if err == services.ErrDuplicateSupplier {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_SUPPLIER",
				"message":    "Supplier dengan nama yang sama sudah ada",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"message":  "Supplier berhasil dibuat",
		"supplier": supplier,
	})
}

// GetSupplier retrieves a supplier by ID
func (h *SupplyChainHandler) GetSupplier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	supplier, err := h.supplierService.GetSupplierByID(uint(id))
	if err != nil {
		if err == services.ErrSupplierNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "SUPPLIER_NOT_FOUND",
				"message":    "Supplier tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"supplier": supplier,
	})
}

// GetAllSuppliers retrieves all suppliers
func (h *SupplyChainHandler) GetAllSuppliers(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"
	query := c.Query("q")
	productCategory := c.Query("product_category")

	var suppliers []models.Supplier
	var err error

	if query != "" || productCategory != "" {
		suppliers, err = h.supplierService.SearchSuppliers(query, productCategory, activeOnly)
	} else {
		suppliers, err = h.supplierService.GetAllSuppliers(activeOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"suppliers": suppliers,
	})
}

// UpdateSupplier updates an existing supplier
func (h *SupplyChainHandler) UpdateSupplier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	supplier := &models.Supplier{
		Name:            req.Name,
		ContactPerson:   req.ContactPerson,
		PhoneNumber:     req.PhoneNumber,
		Email:           req.Email,
		Address:         req.Address,
		ProductCategory: req.ProductCategory,
	}

	if err := h.supplierService.UpdateSupplier(uint(id), supplier); err != nil {
		if err == services.ErrSupplierNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "SUPPLIER_NOT_FOUND",
				"message":    "Supplier tidak ditemukan",
			})
			return
		}

		if err == services.ErrDuplicateSupplier {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_SUPPLIER",
				"message":    "Supplier dengan nama yang sama sudah ada",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Supplier berhasil diperbarui",
	})
}

// GetSupplierPerformance retrieves supplier performance metrics
func (h *SupplyChainHandler) GetSupplierPerformance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	performance, err := h.supplierService.GetSupplierPerformance(uint(id))
	if err != nil {
		if err == services.ErrSupplierNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "SUPPLIER_NOT_FOUND",
				"message":    "Supplier tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"performance": performance,
	})
}

// Purchase Order Endpoints

// CreatePurchaseOrderRequest represents create PO request
type CreatePurchaseOrderRequest struct {
	SupplierID       uint                      `json:"supplier_id" binding:"required"`
	ExpectedDelivery string                    `json:"expected_delivery" binding:"required"`
	Items            []PurchaseOrderItemRequest `json:"items" binding:"required,min=1"`
}

// PurchaseOrderItemRequest represents PO item request
type PurchaseOrderItemRequest struct {
	IngredientID uint    `json:"ingredient_id" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice    float64 `json:"unit_price" binding:"required,gte=0"`
}

// CreatePurchaseOrder creates a new purchase order
func (h *SupplyChainHandler) CreatePurchaseOrder(c *gin.Context) {
	var req CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse expected delivery date
	expectedDelivery, err := time.Parse("2006-01-02", req.ExpectedDelivery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	po := &models.PurchaseOrder{
		SupplierID:       req.SupplierID,
		ExpectedDelivery: expectedDelivery,
	}

	var items []models.PurchaseOrderItem
	for _, item := range req.Items {
		items = append(items, models.PurchaseOrderItem{
			IngredientID: item.IngredientID,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
		})
	}

	if err := h.purchaseOrderService.CreatePurchaseOrder(po, items, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "CREATE_PO_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"message":        "Purchase order berhasil dibuat",
		"purchase_order": po,
	})
}

// GetPurchaseOrder retrieves a purchase order by ID
func (h *SupplyChainHandler) GetPurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	po, err := h.purchaseOrderService.GetPurchaseOrderByID(uint(id))
	if err != nil {
		if err == services.ErrPONotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "PO_NOT_FOUND",
				"message":    "Purchase order tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"purchase_order": po,
	})
}

// GetAllPurchaseOrders retrieves all purchase orders
func (h *SupplyChainHandler) GetAllPurchaseOrders(c *gin.Context) {
	status := c.Query("status")

	pos, err := h.purchaseOrderService.GetAllPurchaseOrders(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"purchase_orders": pos,
	})
}

// UpdatePurchaseOrder updates an existing purchase order
func (h *SupplyChainHandler) UpdatePurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse expected delivery date
	expectedDelivery, err := time.Parse("2006-01-02", req.ExpectedDelivery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	po := &models.PurchaseOrder{
		SupplierID:       req.SupplierID,
		ExpectedDelivery: expectedDelivery,
	}

	var items []models.PurchaseOrderItem
	for _, item := range req.Items {
		items = append(items, models.PurchaseOrderItem{
			IngredientID: item.IngredientID,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
		})
	}

	if err := h.purchaseOrderService.UpdatePurchaseOrder(uint(id), po, items); err != nil {
		if err == services.ErrPONotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "PO_NOT_FOUND",
				"message":    "Purchase order tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "UPDATE_PO_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase order berhasil diperbarui",
	})
}

// ApprovePurchaseOrder approves a purchase order
func (h *SupplyChainHandler) ApprovePurchaseOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	if err := h.purchaseOrderService.ApprovePurchaseOrder(uint(id), userID.(uint)); err != nil {
		if err == services.ErrPONotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "PO_NOT_FOUND",
				"message":    "Purchase order tidak ditemukan",
			})
			return
		}

		if err == services.ErrPOAlreadyApproved {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "PO_ALREADY_APPROVED",
				"message":    "Purchase order sudah disetujui",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "APPROVE_PO_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase order berhasil disetujui",
	})
}

// Goods Receipt Endpoints

// CreateGoodsReceiptRequest represents create GRN request
type CreateGoodsReceiptRequest struct {
	POID  uint                      `json:"po_id" binding:"required"`
	Notes string                    `json:"notes"`
	Items []GoodsReceiptItemRequest `json:"items" binding:"required,min=1"`
}

// GoodsReceiptItemRequest represents GRN item request
type GoodsReceiptItemRequest struct {
	IngredientID     uint    `json:"ingredient_id" binding:"required"`
	ReceivedQuantity float64 `json:"received_quantity" binding:"required,gte=0"`
	ExpiryDate       *string `json:"expiry_date"`
}

// CreateGoodsReceipt creates a new goods receipt
func (h *SupplyChainHandler) CreateGoodsReceipt(c *gin.Context) {
	var req CreateGoodsReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")

	grn := &models.GoodsReceipt{
		POID:  req.POID,
		Notes: req.Notes,
	}

	var items []models.GoodsReceiptItem
	for _, item := range req.Items {
		grnItem := models.GoodsReceiptItem{
			IngredientID:     item.IngredientID,
			ReceivedQuantity: item.ReceivedQuantity,
		}

		// Parse expiry date if provided
		if item.ExpiryDate != nil && *item.ExpiryDate != "" {
			expiryDate, err := time.Parse("2006-01-02", *item.ExpiryDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success":    false,
					"error_code": "INVALID_DATE",
					"message":    "Format tanggal kadaluarsa tidak valid (gunakan YYYY-MM-DD)",
				})
				return
			}
			grnItem.ExpiryDate = &expiryDate
		}

		items = append(items, grnItem)
	}

	if err := h.goodsReceiptService.CreateGoodsReceipt(grn, items, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "CREATE_GRN_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "Goods receipt berhasil dibuat",
		"goods_receipt": grn,
	})
}

// GetAllGoodsReceipts retrieves all goods receipts
func (h *SupplyChainHandler) GetAllGoodsReceipts(c *gin.Context) {
	grns, err := h.goodsReceiptService.GetAllGoodsReceipts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"goods_receipts": grns,
	})
}

// GetGoodsReceipt retrieves a goods receipt by ID
func (h *SupplyChainHandler) GetGoodsReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	grn, err := h.goodsReceiptService.GetGoodsReceiptByID(uint(id))
	if err != nil {
		if err == services.ErrGRNNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "GRN_NOT_FOUND",
				"message":    "Goods receipt tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"goods_receipt": grn,
	})
}

// UploadInvoicePhoto uploads invoice photo for a goods receipt
func (h *SupplyChainHandler) UploadInvoicePhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	// Get file from form
	file, err := c.FormFile("invoice_photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "NO_FILE",
			"message":    "File foto invoice tidak ditemukan",
		})
		return
	}

	// Save file
	filename := fmt.Sprintf("invoice_%d_%d%s", id, time.Now().Unix(), filepath.Ext(file.Filename))
	invoiceDir := filepath.Join(uploadBaseDir, "invoices")
	savePath := filepath.Join(invoiceDir, filename)
	
	// Create directory if not exists
	if err := os.MkdirAll(invoiceDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "UPLOAD_ERROR",
			"message":    "Gagal membuat direktori upload",
		})
		return
	}

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "UPLOAD_ERROR",
			"message":    "Gagal menyimpan file",
		})
		return
	}

	// Generate URL
	photoURL := fmt.Sprintf("/uploads/invoices/%s", filename)

	if err := h.goodsReceiptService.UpdateInvoicePhoto(uint(id), photoURL); err != nil {
		if err == services.ErrGRNNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "GRN_NOT_FOUND",
				"message":    "Goods receipt tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Foto invoice berhasil diunggah",
		"photo_url": photoURL,
	})
}

// Inventory Endpoints

// GetInventory retrieves all inventory items
func (h *SupplyChainHandler) GetInventory(c *gin.Context) {
	items, err := h.inventoryService.GetAllInventory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"inventory_items": items,
	})
}

// GetInventoryAlerts retrieves low stock alerts
func (h *SupplyChainHandler) GetInventoryAlerts(c *gin.Context) {
	alerts, err := h.inventoryService.CheckLowStock()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"alerts":  alerts,
	})
}

// GetInventoryMovements retrieves inventory movements
func (h *SupplyChainHandler) GetInventoryMovements(c *gin.Context) {
	var ingredientID *uint
	if idStr := c.Query("ingredient_id"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err == nil {
			uid := uint(id)
			ingredientID = &uid
		}
	}

	movementType := c.Query("movement_type")

	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if sd, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &sd
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if ed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &ed
		}
	}

	movements, err := h.inventoryService.GetMovements(ingredientID, movementType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"movements": movements,
	})
}

// InitializeInventory initializes inventory records for all ingredients that don't have one
func (h *SupplyChainHandler) InitializeInventory(c *gin.Context) {
	if err := h.inventoryService.InitializeInventoryForAllIngredients(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INITIALIZE_ERROR",
			"message":    "Gagal menginisialisasi inventory: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inventory berhasil diinisialisasi untuk semua bahan",
	})
}

// InitializeInventoryItem initializes inventory record for a specific ingredient
func (h *SupplyChainHandler) InitializeInventoryItem(c *gin.Context) {
	ingredientID, err := strconv.ParseUint(c.Param("ingredient_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID bahan tidak valid",
		})
		return
	}

	if err := h.inventoryService.InitializeInventoryForIngredient(uint(ingredientID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INITIALIZE_ERROR",
			"message":    "Gagal menginisialisasi inventory: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bahan berhasil ditambahkan ke inventory",
	})
}
