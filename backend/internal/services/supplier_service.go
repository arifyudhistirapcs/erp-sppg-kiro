package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrSupplierNotFound   = errors.New("supplier tidak ditemukan")
	ErrSupplierValidation = errors.New("validasi supplier gagal")
	ErrDuplicateSupplier  = errors.New("supplier dengan nama yang sama sudah ada")
)

// SupplierService handles supplier business logic
type SupplierService struct {
	db *gorm.DB
}

// NewSupplierService creates a new supplier service
func NewSupplierService(db *gorm.DB) *SupplierService {
	return &SupplierService{
		db: db,
	}
}

// CreateSupplier creates a new supplier
func (s *SupplierService) CreateSupplier(supplier *models.Supplier) error {
	// Check for duplicate name
	var existing models.Supplier
	err := s.db.Where("name = ?", supplier.Name).First(&existing).Error
	if err == nil {
		return ErrDuplicateSupplier
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Set defaults
	supplier.IsActive = true
	supplier.OnTimeDelivery = 0
	supplier.QualityRating = 0

	return s.db.Create(supplier).Error
}

// GetSupplierByID retrieves a supplier by ID
func (s *SupplierService) GetSupplierByID(id uint) (*models.Supplier, error) {
	var supplier models.Supplier
	err := s.db.First(&supplier, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}

	return &supplier, nil
}

// GetAllSuppliers retrieves all suppliers
func (s *SupplierService) GetAllSuppliers(activeOnly bool) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	query := s.db.Model(&models.Supplier{})
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("name ASC").Find(&suppliers).Error
	return suppliers, err
}

// UpdateSupplier updates an existing supplier
func (s *SupplierService) UpdateSupplier(id uint, updates *models.Supplier) error {
	// Check if supplier exists
	_, err := s.GetSupplierByID(id)
	if err != nil {
		return err
	}

	// Check for duplicate name (excluding current supplier)
	var existing models.Supplier
	err = s.db.Where("name = ? AND id != ?", updates.Name, id).First(&existing).Error
	if err == nil {
		return ErrDuplicateSupplier
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Update supplier
	return s.db.Model(&models.Supplier{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":             updates.Name,
		"contact_person":   updates.ContactPerson,
		"phone_number":     updates.PhoneNumber,
		"email":            updates.Email,
		"address":          updates.Address,
		"product_category": updates.ProductCategory,
		"is_active":        updates.IsActive,
		"updated_at":       time.Now(),
	}).Error
}

// DeactivateSupplier marks a supplier as inactive
func (s *SupplierService) DeactivateSupplier(id uint) error {
	result := s.db.Model(&models.Supplier{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSupplierNotFound
	}
	return nil
}

// ActivateSupplier marks a supplier as active
func (s *SupplierService) ActivateSupplier(id uint) error {
	result := s.db.Model(&models.Supplier{}).Where("id = ?", id).Update("is_active", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSupplierNotFound
	}
	return nil
}

// SupplierPerformance represents performance metrics for a supplier
type SupplierPerformance struct {
	SupplierID       uint      `json:"supplier_id"`
	SupplierName     string    `json:"supplier_name"`
	TotalOrders      int       `json:"total_orders"`
	CompletedOrders  int       `json:"completed_orders"`
	OnTimeDeliveries int       `json:"on_time_deliveries"`
	OnTimeRate       float64   `json:"on_time_rate"`
	QualityRating    float64   `json:"quality_rating"`
	TotalAmount      float64   `json:"total_amount"`
	LastOrderDate    *time.Time `json:"last_order_date"`
}

// GetSupplierPerformance retrieves performance metrics for a supplier
func (s *SupplierService) GetSupplierPerformance(id uint) (*SupplierPerformance, error) {
	supplier, err := s.GetSupplierByID(id)
	if err != nil {
		return nil, err
	}

	performance := &SupplierPerformance{
		SupplierID:   supplier.ID,
		SupplierName: supplier.Name,
		QualityRating: supplier.QualityRating,
	}

	// Count total orders
	var totalOrders int64
	s.db.Model(&models.PurchaseOrder{}).Where("supplier_id = ?", id).Count(&totalOrders)
	performance.TotalOrders = int(totalOrders)

	// Count completed orders (received status)
	var completedOrders int64
	s.db.Model(&models.PurchaseOrder{}).
		Where("supplier_id = ? AND status = ?", id, "received").
		Count(&completedOrders)
	performance.CompletedOrders = int(completedOrders)

	// Calculate on-time deliveries
	// An order is on-time if it was received before or on the expected delivery date
	var onTimeCount int64
	s.db.Model(&models.PurchaseOrder{}).
		Joins("LEFT JOIN goods_receipts ON goods_receipts.po_id = purchase_orders.id").
		Where("purchase_orders.supplier_id = ? AND purchase_orders.status = ? AND goods_receipts.receipt_date <= purchase_orders.expected_delivery", id, "received").
		Count(&onTimeCount)
	performance.OnTimeDeliveries = int(onTimeCount)

	// Calculate on-time rate
	if performance.CompletedOrders > 0 {
		performance.OnTimeRate = float64(performance.OnTimeDeliveries) / float64(performance.CompletedOrders) * 100
	}

	// Calculate total amount
	var totalAmount float64
	s.db.Model(&models.PurchaseOrder{}).
		Where("supplier_id = ? AND status IN ?", id, []string{"approved", "received"}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalAmount)
	performance.TotalAmount = totalAmount

	// Get last order date
	var lastOrder models.PurchaseOrder
	err = s.db.Where("supplier_id = ?", id).
		Order("order_date DESC").
		First(&lastOrder).Error
	if err == nil {
		performance.LastOrderDate = &lastOrder.OrderDate
	}

	// Update supplier's on-time delivery rate
	s.db.Model(&models.Supplier{}).Where("id = ?", id).Update("on_time_delivery", performance.OnTimeRate)

	return performance, nil
}

// UpdateSupplierRating updates the quality rating for a supplier
func (s *SupplierService) UpdateSupplierRating(id uint, rating float64, notes string) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating harus antara 1 dan 5")
	}

	_, err := s.GetSupplierByID(id)
	if err != nil {
		return err
	}

	// Update the rating (simple average for now)
	// In a full implementation, we might want to track individual ratings
	return s.db.Model(&models.Supplier{}).Where("id = ?", id).Update("quality_rating", rating).Error
}

// SearchSuppliers searches suppliers by name or product category
func (s *SupplierService) SearchSuppliers(query string, productCategory string, activeOnly bool) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	db := s.db.Model(&models.Supplier{})

	if activeOnly {
		db = db.Where("is_active = ?", true)
	}

	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}

	if productCategory != "" {
		db = db.Where("product_category = ?", productCategory)
	}

	err := db.Order("name ASC").Find(&suppliers).Error
	return suppliers, err
}
