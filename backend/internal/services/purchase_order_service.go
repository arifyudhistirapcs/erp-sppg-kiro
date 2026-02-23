package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrPONotFound        = errors.New("purchase order tidak ditemukan")
	ErrPOValidation      = errors.New("validasi purchase order gagal")
	ErrPOAlreadyApproved = errors.New("purchase order sudah disetujui")
	ErrPOCancelled       = errors.New("purchase order sudah dibatalkan")
	ErrInvalidPOStatus   = errors.New("status purchase order tidak valid")
)

// PurchaseOrderService handles purchase order business logic
type PurchaseOrderService struct {
	db *gorm.DB
}

// NewPurchaseOrderService creates a new purchase order service
func NewPurchaseOrderService(db *gorm.DB) *PurchaseOrderService {
	return &PurchaseOrderService{
		db: db,
	}
}

// CreatePurchaseOrder creates a new purchase order
func (s *PurchaseOrderService) CreatePurchaseOrder(po *models.PurchaseOrder, items []models.PurchaseOrderItem, userID uint) error {
	// Validate supplier exists and is active
	var supplier models.Supplier
	if err := s.db.First(&supplier, po.SupplierID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("supplier tidak ditemukan")
		}
		return err
	}
	if !supplier.IsActive {
		return errors.New("supplier tidak aktif")
	}

	// Validate items
	if len(items) == 0 {
		return errors.New("purchase order harus memiliki minimal 1 item")
	}

	// Calculate total amount and validate items
	var totalAmount float64
	for i := range items {
		// Validate ingredient exists
		var ingredient models.Ingredient
		if err := s.db.First(&ingredient, items[i].IngredientID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("bahan baku dengan ID %d tidak ditemukan", items[i].IngredientID)
			}
			return err
		}

		// Calculate subtotal
		items[i].Subtotal = items[i].Quantity * items[i].UnitPrice
		totalAmount += items[i].Subtotal
	}

	// Generate PO number
	poNumber, err := s.generatePONumber()
	if err != nil {
		return err
	}

	// Set PO fields
	po.PONumber = poNumber
	po.Status = "pending"
	po.TotalAmount = totalAmount
	po.CreatedBy = userID
	po.OrderDate = time.Now()

	// Create PO in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create PO
		if err := tx.Create(po).Error; err != nil {
			return err
		}

		// Create PO items
		for i := range items {
			items[i].POID = po.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetPurchaseOrderByID retrieves a purchase order by ID with related data
func (s *PurchaseOrderService) GetPurchaseOrderByID(id uint) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	err := s.db.Preload("Supplier").
		Preload("POItems.Ingredient").
		Preload("Creator").
		Preload("Approver").
		First(&po, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPONotFound
		}
		return nil, err
	}

	return &po, nil
}

// GetAllPurchaseOrders retrieves all purchase orders with optional status filter
func (s *PurchaseOrderService) GetAllPurchaseOrders(status string) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	query := s.db.Preload("Supplier").
		Preload("POItems.Ingredient").
		Preload("Creator").
		Preload("Approver")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Order("order_date DESC").Find(&pos).Error
	return pos, err
}

// UpdatePurchaseOrder updates an existing purchase order (only if pending)
func (s *PurchaseOrderService) UpdatePurchaseOrder(id uint, updates *models.PurchaseOrder, items []models.PurchaseOrderItem) error {
	// Get existing PO
	existingPO, err := s.GetPurchaseOrderByID(id)
	if err != nil {
		return err
	}

	// Only allow updates if status is pending
	if existingPO.Status != "pending" {
		return errors.New("hanya purchase order dengan status pending yang dapat diubah")
	}

	// Validate supplier
	var supplier models.Supplier
	if err := s.db.First(&supplier, updates.SupplierID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("supplier tidak ditemukan")
		}
		return err
	}
	if !supplier.IsActive {
		return errors.New("supplier tidak aktif")
	}

	// Validate items
	if len(items) == 0 {
		return errors.New("purchase order harus memiliki minimal 1 item")
	}

	// Calculate total amount
	var totalAmount float64
	for i := range items {
		// Validate ingredient exists
		var ingredient models.Ingredient
		if err := s.db.First(&ingredient, items[i].IngredientID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("bahan baku dengan ID %d tidak ditemukan", items[i].IngredientID)
			}
			return err
		}

		// Calculate subtotal
		items[i].Subtotal = items[i].Quantity * items[i].UnitPrice
		totalAmount += items[i].Subtotal
	}

	// Update PO in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete old PO items
		if err := tx.Where("po_id = ?", id).Delete(&models.PurchaseOrderItem{}).Error; err != nil {
			return err
		}

		// Update PO
		if err := tx.Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
			"supplier_id":       updates.SupplierID,
			"expected_delivery": updates.ExpectedDelivery,
			"total_amount":      totalAmount,
			"updated_at":        time.Now(),
		}).Error; err != nil {
			return err
		}

		// Create new PO items
		for i := range items {
			items[i].POID = id
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}

// ApprovePurchaseOrder approves a purchase order
func (s *PurchaseOrderService) ApprovePurchaseOrder(id uint, approverID uint) error {
	// Get existing PO
	po, err := s.GetPurchaseOrderByID(id)
	if err != nil {
		return err
	}

	// Check if already approved
	if po.Status == "approved" {
		return ErrPOAlreadyApproved
	}

	// Check if cancelled
	if po.Status == "cancelled" {
		return ErrPOCancelled
	}

	// Only allow approval if status is pending
	if po.Status != "pending" {
		return ErrInvalidPOStatus
	}

	// Update status to approved
	now := time.Now()
	return s.db.Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "approved",
		"approved_by": approverID,
		"approved_at": now,
		"updated_at":  now,
	}).Error
}

// CancelPurchaseOrder cancels a purchase order
func (s *PurchaseOrderService) CancelPurchaseOrder(id uint) error {
	// Get existing PO
	po, err := s.GetPurchaseOrderByID(id)
	if err != nil {
		return err
	}

	// Check if already received
	if po.Status == "received" {
		return errors.New("purchase order yang sudah diterima tidak dapat dibatalkan")
	}

	// Update status to cancelled
	return s.db.Model(&models.PurchaseOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "cancelled",
		"updated_at": time.Now(),
	}).Error
}

// GetPendingPurchaseOrders retrieves all pending purchase orders
func (s *PurchaseOrderService) GetPendingPurchaseOrders() ([]models.PurchaseOrder, error) {
	return s.GetAllPurchaseOrders("pending")
}

// GetApprovedPurchaseOrders retrieves all approved purchase orders
func (s *PurchaseOrderService) GetApprovedPurchaseOrders() ([]models.PurchaseOrder, error) {
	return s.GetAllPurchaseOrders("approved")
}

// GetPurchaseOrdersBySupplier retrieves all purchase orders for a specific supplier
func (s *PurchaseOrderService) GetPurchaseOrdersBySupplier(supplierID uint) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	err := s.db.Preload("Supplier").
		Preload("POItems.Ingredient").
		Preload("Creator").
		Preload("Approver").
		Where("supplier_id = ?", supplierID).
		Order("order_date DESC").
		Find(&pos).Error
	return pos, err
}

// GetPurchaseOrdersByDateRange retrieves purchase orders within a date range
func (s *PurchaseOrderService) GetPurchaseOrdersByDateRange(startDate, endDate time.Time) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	err := s.db.Preload("Supplier").
		Preload("POItems.Ingredient").
		Preload("Creator").
		Preload("Approver").
		Where("order_date BETWEEN ? AND ?", startDate, endDate).
		Order("order_date DESC").
		Find(&pos).Error
	return pos, err
}

// generatePONumber generates a unique PO number
func (s *PurchaseOrderService) generatePONumber() (string, error) {
	// Format: PO-YYYYMMDD-XXXX
	now := time.Now()
	datePrefix := now.Format("20060102")
	
	// Count POs created today
	var count int64
	s.db.Model(&models.PurchaseOrder{}).
		Where("po_number LIKE ?", fmt.Sprintf("PO-%s-%%", datePrefix)).
		Count(&count)
	
	// Generate PO number
	poNumber := fmt.Sprintf("PO-%s-%04d", datePrefix, count+1)
	
	// Check if it already exists (race condition protection)
	var existing models.PurchaseOrder
	err := s.db.Where("po_number = ?", poNumber).First(&existing).Error
	if err == nil {
		// If exists, try with incremented number
		poNumber = fmt.Sprintf("PO-%s-%04d", datePrefix, count+2)
	}
	
	return poNumber, nil
}
