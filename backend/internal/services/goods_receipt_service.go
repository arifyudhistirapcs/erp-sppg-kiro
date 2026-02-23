package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrGRNNotFound       = errors.New("goods receipt tidak ditemukan")
	ErrGRNValidation     = errors.New("validasi goods receipt gagal")
	ErrPONotApproved     = errors.New("purchase order belum disetujui")
	ErrPOAlreadyReceived = errors.New("purchase order sudah diterima")
)

// GoodsReceiptService handles goods receipt business logic
type GoodsReceiptService struct {
	db               *gorm.DB
	inventoryService *InventoryService
	cashFlowService  *CashFlowService
}

// NewGoodsReceiptService creates a new goods receipt service
func NewGoodsReceiptService(db *gorm.DB, inventoryService *InventoryService, cashFlowService *CashFlowService) *GoodsReceiptService {
	return &GoodsReceiptService{
		db:               db,
		inventoryService: inventoryService,
		cashFlowService:  cashFlowService,
	}
}

// QuantityDiscrepancy represents a discrepancy between ordered and received quantities
type QuantityDiscrepancy struct {
	IngredientID     uint    `json:"ingredient_id"`
	IngredientName   string  `json:"ingredient_name"`
	OrderedQuantity  float64 `json:"ordered_quantity"`
	ReceivedQuantity float64 `json:"received_quantity"`
	Difference       float64 `json:"difference"`
	DifferencePercent float64 `json:"difference_percent"`
}

// CreateGoodsReceipt creates a new goods receipt and updates inventory
func (s *GoodsReceiptService) CreateGoodsReceipt(grn *models.GoodsReceipt, items []models.GoodsReceiptItem, userID uint) error {
	// Validate PO exists and is approved
	var po models.PurchaseOrder
	if err := s.db.Preload("POItems.Ingredient").First(&po, grn.POID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("purchase order tidak ditemukan")
		}
		return err
	}

	if po.Status != "approved" {
		return ErrPONotApproved
	}

	if po.Status == "received" {
		return ErrPOAlreadyReceived
	}

	// Validate items
	if len(items) == 0 {
		return errors.New("goods receipt harus memiliki minimal 1 item")
	}

	// Validate items match PO items
	poItemsMap := make(map[uint]*models.PurchaseOrderItem)
	for i := range po.POItems {
		poItemsMap[po.POItems[i].IngredientID] = &po.POItems[i]
	}

	for i := range items {
		poItem, exists := poItemsMap[items[i].IngredientID]
		if !exists {
			return fmt.Errorf("bahan baku dengan ID %d tidak ada dalam purchase order", items[i].IngredientID)
		}
		items[i].OrderedQuantity = poItem.Quantity
	}

	// Generate GRN number
	grnNumber, err := s.generateGRNNumber()
	if err != nil {
		return err
	}

	// Set GRN fields
	grn.GRNNumber = grnNumber
	grn.ReceivedBy = userID
	grn.ReceiptDate = time.Now()

	// Create GRN in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create GRN
		if err := tx.Create(grn).Error; err != nil {
			return err
		}

		// Create GRN items
		for i := range items {
			items[i].GRNID = grn.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		// Update PO status to received
		if err := tx.Model(&models.PurchaseOrder{}).Where("id = ?", grn.POID).Updates(map[string]interface{}{
			"status":     "received",
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		// Update inventory for each item
		for _, item := range items {
			if err := s.inventoryService.UpdateStockWithTx(tx, item.IngredientID, item.ReceivedQuantity, "in", grn.GRNNumber, userID, ""); err != nil {
				return err
			}
		}

		// Create cash flow entry
		if s.cashFlowService != nil {
			cashFlowEntry := &models.CashFlowEntry{
				Date:        grn.ReceiptDate,
				Category:    "bahan_baku",
				Type:        "expense",
				Amount:      po.TotalAmount,
				Description: fmt.Sprintf("Pembelian bahan baku dari %s (PO: %s)", po.Supplier.Name, po.PONumber),
				Reference:   grn.GRNNumber,
				CreatedBy:   userID,
			}
			if err := s.cashFlowService.CreateCashFlowEntryWithTx(tx, cashFlowEntry); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetGoodsReceiptByID retrieves a goods receipt by ID with related data
func (s *GoodsReceiptService) GetGoodsReceiptByID(id uint) (*models.GoodsReceipt, error) {
	var grn models.GoodsReceipt
	err := s.db.Preload("PurchaseOrder.Supplier").
		Preload("PurchaseOrder.POItems.Ingredient").
		Preload("GRNItems.Ingredient").
		Preload("Receiver").
		First(&grn, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGRNNotFound
		}
		return nil, err
	}

	return &grn, nil
}

// GetAllGoodsReceipts retrieves all goods receipts
func (s *GoodsReceiptService) GetAllGoodsReceipts() ([]models.GoodsReceipt, error) {
	var grns []models.GoodsReceipt
	err := s.db.Preload("PurchaseOrder.Supplier").
		Preload("GRNItems.Ingredient").
		Preload("Receiver").
		Order("receipt_date DESC").
		Find(&grns).Error
	return grns, err
}

// GetGoodsReceiptsByDateRange retrieves goods receipts within a date range
func (s *GoodsReceiptService) GetGoodsReceiptsByDateRange(startDate, endDate time.Time) ([]models.GoodsReceipt, error) {
	var grns []models.GoodsReceipt
	err := s.db.Preload("PurchaseOrder.Supplier").
		Preload("GRNItems.Ingredient").
		Preload("Receiver").
		Where("receipt_date BETWEEN ? AND ?", startDate, endDate).
		Order("receipt_date DESC").
		Find(&grns).Error
	return grns, err
}

// GetGoodsReceiptsByPO retrieves goods receipt for a specific purchase order
func (s *GoodsReceiptService) GetGoodsReceiptsByPO(poID uint) (*models.GoodsReceipt, error) {
	var grn models.GoodsReceipt
	err := s.db.Preload("PurchaseOrder.Supplier").
		Preload("GRNItems.Ingredient").
		Preload("Receiver").
		Where("po_id = ?", poID).
		First(&grn).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGRNNotFound
		}
		return nil, err
	}

	return &grn, nil
}

// CheckQuantityDiscrepancies checks for discrepancies between ordered and received quantities
func (s *GoodsReceiptService) CheckQuantityDiscrepancies(grnID uint) ([]QuantityDiscrepancy, error) {
	grn, err := s.GetGoodsReceiptByID(grnID)
	if err != nil {
		return nil, err
	}

	var discrepancies []QuantityDiscrepancy
	for _, item := range grn.GRNItems {
		if item.OrderedQuantity != item.ReceivedQuantity {
			diff := item.ReceivedQuantity - item.OrderedQuantity
			diffPercent := 0.0
			if item.OrderedQuantity > 0 {
				diffPercent = (diff / item.OrderedQuantity) * 100
			}

			discrepancies = append(discrepancies, QuantityDiscrepancy{
				IngredientID:      item.IngredientID,
				IngredientName:    item.Ingredient.Name,
				OrderedQuantity:   item.OrderedQuantity,
				ReceivedQuantity:  item.ReceivedQuantity,
				Difference:        diff,
				DifferencePercent: diffPercent,
			})
		}
	}

	return discrepancies, nil
}

// UpdateInvoicePhoto updates the invoice photo URL for a goods receipt
func (s *GoodsReceiptService) UpdateInvoicePhoto(grnID uint, photoURL string) error {
	result := s.db.Model(&models.GoodsReceipt{}).Where("id = ?", grnID).Update("invoice_photo", photoURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrGRNNotFound
	}
	return nil
}

// generateGRNNumber generates a unique GRN number
func (s *GoodsReceiptService) generateGRNNumber() (string, error) {
	// Format: GRN-YYYYMMDD-XXXX
	now := time.Now()
	datePrefix := now.Format("20060102")
	
	// Count GRNs created today
	var count int64
	s.db.Model(&models.GoodsReceipt{}).
		Where("grn_number LIKE ?", fmt.Sprintf("GRN-%s-%%", datePrefix)).
		Count(&count)
	
	// Generate GRN number
	grnNumber := fmt.Sprintf("GRN-%s-%04d", datePrefix, count+1)
	
	// Check if it already exists (race condition protection)
	var existing models.GoodsReceipt
	err := s.db.Where("grn_number = ?", grnNumber).First(&existing).Error
	if err == nil {
		// If exists, try with incremented number
		grnNumber = fmt.Sprintf("GRN-%s-%04d", datePrefix, count+2)
	}
	
	return grnNumber, nil
}
