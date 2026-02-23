package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FinancialHandler handles financial endpoints
type FinancialHandler struct {
	assetService           *services.AssetService
	cashFlowService        *services.CashFlowService
	financialReportService *services.FinancialReportService
}

// NewFinancialHandler creates a new financial handler
func NewFinancialHandler(db *gorm.DB) *FinancialHandler {
	return &FinancialHandler{
		assetService:           services.NewAssetService(db),
		cashFlowService:        services.NewCashFlowService(db),
		financialReportService: services.NewFinancialReportService(db),
	}
}

// Asset Endpoints

// CreateAssetRequest represents create asset request
type CreateAssetRequest struct {
	AssetCode        string  `json:"asset_code" binding:"required"`
	Name             string  `json:"name" binding:"required"`
	Category         string  `json:"category"`
	PurchaseDate     string  `json:"purchase_date" binding:"required"`
	PurchasePrice    float64 `json:"purchase_price" binding:"required,gte=0"`
	DepreciationRate float64 `json:"depreciation_rate" binding:"gte=0,lte=100"`
	Condition        string  `json:"condition" binding:"required,oneof=good fair poor"`
	Location         string  `json:"location"`
}

// CreateAsset creates a new kitchen asset
func (h *FinancialHandler) CreateAsset(c *gin.Context) {
	var req CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse purchase date
	purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	asset := &models.KitchenAsset{
		AssetCode:        req.AssetCode,
		Name:             req.Name,
		Category:         req.Category,
		PurchaseDate:     purchaseDate,
		PurchasePrice:    req.PurchasePrice,
		DepreciationRate: req.DepreciationRate,
		Condition:        req.Condition,
		Location:         req.Location,
	}

	if err := h.assetService.CreateAsset(asset); err != nil {
		if err == services.ErrDuplicateAssetCode {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_ASSET_CODE",
				"message":    "Kode aset sudah digunakan",
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
		"success": true,
		"message": "Aset berhasil dibuat",
		"asset":   asset,
	})
}

// GetAsset retrieves an asset by ID
func (h *FinancialHandler) GetAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	asset, err := h.assetService.GetAssetByID(uint(id))
	if err != nil {
		if err == services.ErrAssetNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ASSET_NOT_FOUND",
				"message":    "Aset tidak ditemukan",
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
		"asset":   asset,
	})
}

// GetAllAssets retrieves all assets
func (h *FinancialHandler) GetAllAssets(c *gin.Context) {
	category := c.Query("category")
	query := c.Query("q")
	condition := c.Query("condition")

	var assets []models.KitchenAsset
	var err error

	if query != "" || condition != "" {
		assets, err = h.assetService.SearchAssets(query, category, condition)
	} else {
		assets, err = h.assetService.GetAllAssets(category)
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
		"success": true,
		"assets":  assets,
	})
}

// UpdateAsset updates an existing asset
func (h *FinancialHandler) UpdateAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse purchase date
	purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	asset := &models.KitchenAsset{
		AssetCode:        req.AssetCode,
		Name:             req.Name,
		Category:         req.Category,
		PurchaseDate:     purchaseDate,
		PurchasePrice:    req.PurchasePrice,
		DepreciationRate: req.DepreciationRate,
		Condition:        req.Condition,
		Location:         req.Location,
	}

	if err := h.assetService.UpdateAsset(uint(id), asset); err != nil {
		if err == services.ErrAssetNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ASSET_NOT_FOUND",
				"message":    "Aset tidak ditemukan",
			})
			return
		}

		if err == services.ErrDuplicateAssetCode {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "DUPLICATE_ASSET_CODE",
				"message":    "Kode aset sudah digunakan",
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
		"message": "Aset berhasil diperbarui",
	})
}

// DeleteAsset deletes an asset
func (h *FinancialHandler) DeleteAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	if err := h.assetService.DeleteAsset(uint(id)); err != nil {
		if err == services.ErrAssetNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ASSET_NOT_FOUND",
				"message":    "Aset tidak ditemukan",
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
		"message": "Aset berhasil dihapus",
	})
}

// AddMaintenanceRequest represents add maintenance request
type AddMaintenanceRequest struct {
	MaintenanceDate string  `json:"maintenance_date" binding:"required"`
	Description     string  `json:"description"`
	Cost            float64 `json:"cost" binding:"gte=0"`
	PerformedBy     string  `json:"performed_by"`
}

// AddMaintenance adds a maintenance record for an asset
func (h *FinancialHandler) AddMaintenance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	var req AddMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	// Parse maintenance date
	maintenanceDate, err := time.Parse("2006-01-02", req.MaintenanceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	maintenance := &models.AssetMaintenance{
		MaintenanceDate: maintenanceDate,
		Description:     req.Description,
		Cost:            req.Cost,
		PerformedBy:     req.PerformedBy,
	}

	if err := h.assetService.AddMaintenanceRecord(uint(id), maintenance); err != nil {
		if err == services.ErrAssetNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ASSET_NOT_FOUND",
				"message":    "Aset tidak ditemukan",
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
		"success":     true,
		"message":     "Catatan maintenance berhasil ditambahkan",
		"maintenance": maintenance,
	})
}

// GetAssetReport retrieves asset report
func (h *FinancialHandler) GetAssetReport(c *gin.Context) {
	report, err := h.assetService.GenerateAssetReport()
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
		"report":  report,
	})
}

// GetDepreciationSchedule retrieves depreciation schedule for an asset
func (h *FinancialHandler) GetDepreciationSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_ID",
			"message":    "ID tidak valid",
		})
		return
	}

	years := 5 // Default
	if yearsStr := c.Query("years"); yearsStr != "" {
		if y, err := strconv.Atoi(yearsStr); err == nil && y > 0 {
			years = y
		}
	}

	schedule, err := h.assetService.GetDepreciationSchedule(uint(id), years)
	if err != nil {
		if err == services.ErrAssetNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error_code": "ASSET_NOT_FOUND",
				"message":    "Aset tidak ditemukan",
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
		"schedule": schedule,
	})
}

// Cash Flow Endpoints

// CreateCashFlowRequest represents create cash flow request
type CreateCashFlowRequest struct {
	Date        string  `json:"date"`
	Category    string  `json:"category" binding:"required,oneof=bahan_baku gaji utilitas operasional"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Reference   string  `json:"reference"`
}

// CreateCashFlow creates a new cash flow entry
func (h *FinancialHandler) CreateCashFlow(c *gin.Context) {
	var req CreateCashFlowRequest
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

	// Parse date if provided
	var entryDate time.Time
	if req.Date != "" {
		var err error
		entryDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error_code": "INVALID_DATE",
				"message":    "Format tanggal tidak valid (gunakan YYYY-MM-DD)",
			})
			return
		}
	} else {
		entryDate = time.Now()
	}

	entry := &models.CashFlowEntry{
		Date:        entryDate,
		Category:    req.Category,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Reference:   req.Reference,
		CreatedBy:   userID.(uint),
	}

	if err := h.cashFlowService.CreateCashFlowEntry(entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "CREATE_CASH_FLOW_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"message":    "Cash flow entry berhasil dibuat",
		"cash_flow": entry,
	})
}

// GetAllCashFlow retrieves all cash flow entries
func (h *FinancialHandler) GetAllCashFlow(c *gin.Context) {
	category := c.Query("category")
	entryType := c.Query("type")

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

	entries, err := h.cashFlowService.GetAllCashFlowEntries(category, entryType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "INTERNAL_ERROR",
			"message":    "Terjadi kesalahan pada server",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"cash_flows": entries,
	})
}

// GetCashFlowSummary retrieves cash flow summary
func (h *FinancialHandler) GetCashFlowSummary(c *gin.Context) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "MISSING_DATES",
			"message":    "Parameter start_date dan end_date diperlukan",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format start_date tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format end_date tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	summary, err := h.cashFlowService.GetCashFlowSummary(startDate, endDate)
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
		"summary": summary,
	})
}

// Financial Report Endpoints

// GetFinancialReport generates a financial report
func (h *FinancialHandler) GetFinancialReport(c *gin.Context) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "MISSING_DATES",
			"message":    "Parameter start_date dan end_date diperlukan",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format start_date tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format end_date tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	// Parse options
	includeBudget := c.DefaultQuery("include_budget", "false") == "true"
	includeAssets := c.DefaultQuery("include_assets", "false") == "true"
	includeTrend := c.DefaultQuery("include_trend", "false") == "true"

	report, err := h.financialReportService.GenerateFinancialReport(
		startDate,
		endDate,
		includeBudget,
		includeAssets,
		includeTrend,
	)
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
		"report":  report,
	})
}

// ExportFinancialReportRequest represents export request
type ExportFinancialReportRequest struct {
	StartDate      string `json:"start_date" binding:"required"`
	EndDate        string `json:"end_date" binding:"required"`
	Format         string `json:"format" binding:"required,oneof=pdf excel"`
	IncludeBudget  bool   `json:"include_budget"`
	IncludeAssets  bool   `json:"include_assets"`
	IncludeTrend   bool   `json:"include_trend"`
	IncludeCharts  bool   `json:"include_charts"`
}

// ExportFinancialReport exports a financial report
func (h *FinancialHandler) ExportFinancialReport(c *gin.Context) {
	var req ExportFinancialReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "VALIDATION_ERROR",
			"message":    "Data tidak valid",
			"details":    err.Error(),
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format start_date tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error_code": "INVALID_DATE",
			"message":    "Format end_date tidak valid (gunakan YYYY-MM-DD)",
		})
		return
	}

	options := services.ExportOptions{
		Format:        services.ExportFormat(req.Format),
		IncludeBudget: req.IncludeBudget,
		IncludeAssets: req.IncludeAssets,
		IncludeTrend:  req.IncludeTrend,
		IncludeCharts: req.IncludeCharts,
	}

	data, filename, err := h.financialReportService.ExportFinancialReport(startDate, endDate, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error_code": "EXPORT_ERROR",
			"message":    err.Error(),
		})
		return
	}

	// Set appropriate content type
	contentType := "application/pdf"
	if req.Format == "excel" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}
