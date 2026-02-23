package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrInventoryNotFound    = errors.New("inventory item tidak ditemukan")
	ErrInsufficientStock    = errors.New("stok tidak mencukupi")
	ErrInvalidMovementType  = errors.New("tipe pergerakan inventory tidak valid")
)

// InventoryService handles inventory business logic
type InventoryService struct {
	db *gorm.DB
}

// NewInventoryService creates a new inventory service
func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{
		db: db,
	}
}

// UpdateStock updates inventory stock levels and creates movement record
func (s *InventoryService) UpdateStock(ingredientID uint, quantity float64, movementType string, reference string, userID uint, notes string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.UpdateStockWithTx(tx, ingredientID, quantity, movementType, reference, userID, notes)
	})
}

// UpdateStockWithTx updates inventory stock levels within a transaction
func (s *InventoryService) UpdateStockWithTx(tx *gorm.DB, ingredientID uint, quantity float64, movementType string, reference string, userID uint, notes string) error {
	// Validate movement type
	if movementType != "in" && movementType != "out" && movementType != "adjustment" {
		return ErrInvalidMovementType
	}

	// Get or create inventory item
	var inventoryItem models.InventoryItem
	err := tx.Where("ingredient_id = ?", ingredientID).First(&inventoryItem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new inventory item
			inventoryItem = models.InventoryItem{
				IngredientID: ingredientID,
				Quantity:     0,
				MinThreshold: 10, // default threshold
				LastUpdated:  time.Now(),
			}
			if err := tx.Create(&inventoryItem).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Calculate new quantity
	var newQuantity float64
	switch movementType {
	case "in", "adjustment":
		newQuantity = inventoryItem.Quantity + quantity
	case "out":
		newQuantity = inventoryItem.Quantity - quantity
		if newQuantity < 0 {
			return ErrInsufficientStock
		}
	}

	// Update inventory quantity
	if err := tx.Model(&models.InventoryItem{}).Where("id = ?", inventoryItem.ID).Updates(map[string]interface{}{
		"quantity":     newQuantity,
		"last_updated": time.Now(),
	}).Error; err != nil {
		return err
	}

	// Create inventory movement record
	movement := models.InventoryMovement{
		IngredientID: ingredientID,
		MovementType: movementType,
		Quantity:     quantity,
		Reference:    reference,
		MovementDate: time.Now(),
		CreatedBy:    userID,
		Notes:        notes,
	}
	if err := tx.Create(&movement).Error; err != nil {
		return err
	}

	return nil
}

// GetInventoryItem retrieves inventory item by ingredient ID
func (s *InventoryService) GetInventoryItem(ingredientID uint) (*models.InventoryItem, error) {
	var item models.InventoryItem
	err := s.db.Preload("Ingredient").Where("ingredient_id = ?", ingredientID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}
		return nil, err
	}
	return &item, nil
}

// GetAllInventory retrieves all inventory items
func (s *InventoryService) GetAllInventory() ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	err := s.db.Preload("Ingredient").Order("ingredient_id ASC").Find(&items).Error
	return items, err
}

// LowStockAlert represents a low stock alert
type LowStockAlert struct {
	IngredientID   uint    `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	CurrentStock   float64 `json:"current_stock"`
	MinThreshold   float64 `json:"min_threshold"`
	Unit           string  `json:"unit"`
	DaysRemaining  float64 `json:"days_remaining"`
}

// CheckLowStock checks for items below minimum threshold
func (s *InventoryService) CheckLowStock() ([]LowStockAlert, error) {
	var items []models.InventoryItem
	err := s.db.Preload("Ingredient").
		Where("quantity < min_threshold").
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	var alerts []LowStockAlert
	for _, item := range items {
		// Calculate days remaining based on average daily consumption
		daysRemaining := s.calculateDaysRemaining(item.IngredientID, item.Quantity)

		alerts = append(alerts, LowStockAlert{
			IngredientID:   item.IngredientID,
			IngredientName: item.Ingredient.Name,
			CurrentStock:   item.Quantity,
			MinThreshold:   item.MinThreshold,
			Unit:           item.Ingredient.Unit,
			DaysRemaining:  daysRemaining,
		})
	}

	return alerts, nil
}

// calculateDaysRemaining calculates estimated days of supply remaining
func (s *InventoryService) calculateDaysRemaining(ingredientID uint, currentStock float64) float64 {
	// Calculate average daily consumption over last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	var totalOut float64
	s.db.Model(&models.InventoryMovement{}).
		Where("ingredient_id = ? AND movement_type = ? AND movement_date >= ?", ingredientID, "out", thirtyDaysAgo).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalOut)

	if totalOut == 0 {
		return 999 // No consumption data, return high number
	}

	avgDailyConsumption := totalOut / 30.0
	if avgDailyConsumption == 0 {
		return 999
	}

	return currentStock / avgDailyConsumption
}

// GetMovements retrieves inventory movements with optional filters
func (s *InventoryService) GetMovements(ingredientID *uint, movementType string, startDate, endDate *time.Time) ([]models.InventoryMovement, error) {
	query := s.db.Preload("Ingredient").Preload("Creator")

	if ingredientID != nil {
		query = query.Where("ingredient_id = ?", *ingredientID)
	}

	if movementType != "" {
		query = query.Where("movement_type = ?", movementType)
	}

	if startDate != nil && endDate != nil {
		query = query.Where("movement_date BETWEEN ? AND ?", *startDate, *endDate)
	}

	var movements []models.InventoryMovement
	err := query.Order("movement_date DESC").Find(&movements).Error
	return movements, err
}

// UpdateMinThreshold updates the minimum threshold for an ingredient
func (s *InventoryService) UpdateMinThreshold(ingredientID uint, threshold float64) error {
	result := s.db.Model(&models.InventoryItem{}).
		Where("ingredient_id = ?", ingredientID).
		Update("min_threshold", threshold)
	
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInventoryNotFound
	}
	return nil
}

// InventoryBatch represents a batch of inventory with expiry date for FIFO/FEFO
type InventoryBatch struct {
	IngredientID uint
	Quantity     float64
	ExpiryDate   *time.Time
	ReceiptDate  time.Time
	Reference    string
}

// GetInventoryBatches retrieves inventory batches for FIFO/FEFO processing
func (s *InventoryService) GetInventoryBatches(ingredientID uint) ([]InventoryBatch, error) {
	// Get all "in" movements for this ingredient that haven't been fully consumed
	var movements []models.InventoryMovement
	err := s.db.Where("ingredient_id = ? AND movement_type = ?", ingredientID, "in").
		Order("movement_date ASC").
		Find(&movements).Error
	if err != nil {
		return nil, err
	}

	var batches []InventoryBatch
	for _, movement := range movements {
		// Get expiry date from GRN if available
		var expiryDate *time.Time
		if movement.Reference != "" {
			var grnItem models.GoodsReceiptItem
			err := s.db.Joins("JOIN goods_receipts ON goods_receipts.id = goods_receipt_items.grn_id").
				Where("goods_receipts.grn_number = ? AND goods_receipt_items.ingredient_id = ?", movement.Reference, ingredientID).
				First(&grnItem).Error
			if err == nil {
				expiryDate = grnItem.ExpiryDate
			}
		}

		batches = append(batches, InventoryBatch{
			IngredientID: ingredientID,
			Quantity:     movement.Quantity,
			ExpiryDate:   expiryDate,
			ReceiptDate:  movement.MovementDate,
			Reference:    movement.Reference,
		})
	}

	return batches, nil
}

// ConsumeBatchesFIFO consumes inventory using FIFO method
func (s *InventoryService) ConsumeBatchesFIFO(ingredientID uint, quantityNeeded float64) ([]InventoryBatch, error) {
	batches, err := s.GetInventoryBatches(ingredientID)
	if err != nil {
		return nil, err
	}

	// Sort by receipt date (FIFO)
	// Already sorted by movement_date ASC in GetInventoryBatches

	return s.consumeBatches(batches, quantityNeeded)
}

// ConsumeBatchesFEFO consumes inventory using FEFO method
func (s *InventoryService) ConsumeBatchesFEFO(ingredientID uint, quantityNeeded float64) ([]InventoryBatch, error) {
	batches, err := s.GetInventoryBatches(ingredientID)
	if err != nil {
		return nil, err
	}

	// Sort by expiry date (FEFO) - items with expiry date first, then by date
	// Items without expiry date go to the end
	sortedBatches := make([]InventoryBatch, 0, len(batches))
	var noExpiryBatches []InventoryBatch

	for _, batch := range batches {
		if batch.ExpiryDate != nil {
			sortedBatches = append(sortedBatches, batch)
		} else {
			noExpiryBatches = append(noExpiryBatches, batch)
		}
	}

	// Sort batches with expiry date by expiry date
	for i := 0; i < len(sortedBatches)-1; i++ {
		for j := i + 1; j < len(sortedBatches); j++ {
			if sortedBatches[i].ExpiryDate.After(*sortedBatches[j].ExpiryDate) {
				sortedBatches[i], sortedBatches[j] = sortedBatches[j], sortedBatches[i]
			}
		}
	}

	// Append batches without expiry date at the end
	sortedBatches = append(sortedBatches, noExpiryBatches...)

	return s.consumeBatches(sortedBatches, quantityNeeded)
}

// consumeBatches consumes batches in order until quantity is satisfied
func (s *InventoryService) consumeBatches(batches []InventoryBatch, quantityNeeded float64) ([]InventoryBatch, error) {
	var consumedBatches []InventoryBatch
	remainingNeeded := quantityNeeded

	for _, batch := range batches {
		if remainingNeeded <= 0 {
			break
		}

		if batch.Quantity >= remainingNeeded {
			// This batch can fulfill the remaining need
			consumedBatches = append(consumedBatches, InventoryBatch{
				IngredientID: batch.IngredientID,
				Quantity:     remainingNeeded,
				ExpiryDate:   batch.ExpiryDate,
				ReceiptDate:  batch.ReceiptDate,
				Reference:    batch.Reference,
			})
			remainingNeeded = 0
		} else {
			// Consume entire batch and continue
			consumedBatches = append(consumedBatches, batch)
			remainingNeeded -= batch.Quantity
		}
	}

	if remainingNeeded > 0 {
		return nil, fmt.Errorf("stok tidak mencukupi: masih membutuhkan %.2f", remainingNeeded)
	}

	return consumedBatches, nil
}

// GetStockReport generates a stock report for a date range
func (s *InventoryService) GetStockReport(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var items []models.InventoryItem
	err := s.db.Preload("Ingredient").Find(&items).Error
	if err != nil {
		return nil, err
	}

	var report []map[string]interface{}
	for _, item := range items {
		// Get movements for this ingredient in date range
		var inQuantity, outQuantity float64
		
		s.db.Model(&models.InventoryMovement{}).
			Where("ingredient_id = ? AND movement_type = ? AND movement_date BETWEEN ? AND ?", 
				item.IngredientID, "in", startDate, endDate).
			Select("COALESCE(SUM(quantity), 0)").
			Scan(&inQuantity)

		s.db.Model(&models.InventoryMovement{}).
			Where("ingredient_id = ? AND movement_type = ? AND movement_date BETWEEN ? AND ?", 
				item.IngredientID, "out", startDate, endDate).
			Select("COALESCE(SUM(quantity), 0)").
			Scan(&outQuantity)

		report = append(report, map[string]interface{}{
			"ingredient_id":   item.IngredientID,
			"ingredient_name": item.Ingredient.Name,
			"unit":            item.Ingredient.Unit,
			"current_stock":   item.Quantity,
			"min_threshold":   item.MinThreshold,
			"stock_in":        inQuantity,
			"stock_out":       outQuantity,
			"net_change":      inQuantity - outQuantity,
		})
	}

	return report, nil
}
