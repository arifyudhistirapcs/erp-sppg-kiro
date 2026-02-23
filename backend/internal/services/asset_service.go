package services

import (
	"errors"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrAssetNotFound      = errors.New("aset tidak ditemukan")
	ErrAssetValidation    = errors.New("validasi aset gagal")
	ErrDuplicateAssetCode = errors.New("kode aset sudah digunakan")
)

// AssetService handles kitchen asset business logic
type AssetService struct {
	db *gorm.DB
}

// NewAssetService creates a new asset service
func NewAssetService(db *gorm.DB) *AssetService {
	return &AssetService{
		db: db,
	}
}

// CreateAsset creates a new kitchen asset
func (s *AssetService) CreateAsset(asset *models.KitchenAsset) error {
	// Check for duplicate asset code
	var existing models.KitchenAsset
	err := s.db.Where("asset_code = ?", asset.AssetCode).First(&existing).Error
	if err == nil {
		return ErrDuplicateAssetCode
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Set initial current value to purchase price
	asset.CurrentValue = asset.PurchasePrice

	// Calculate initial depreciation if purchase date is in the past
	if asset.PurchaseDate.Before(time.Now()) {
		asset.CurrentValue = s.CalculateBookValue(asset.PurchasePrice, asset.PurchaseDate, asset.DepreciationRate)
	}

	return s.db.Create(asset).Error
}

// GetAssetByID retrieves an asset by ID with maintenance records
func (s *AssetService) GetAssetByID(id uint) (*models.KitchenAsset, error) {
	var asset models.KitchenAsset
	err := s.db.Preload("MaintenanceRecords").First(&asset, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	// Recalculate current value based on depreciation
	asset.CurrentValue = s.CalculateBookValue(asset.PurchasePrice, asset.PurchaseDate, asset.DepreciationRate)

	return &asset, nil
}

// GetAllAssets retrieves all assets
func (s *AssetService) GetAllAssets(category string) ([]models.KitchenAsset, error) {
	var assets []models.KitchenAsset
	query := s.db.Model(&models.KitchenAsset{})
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	err := query.Order("name ASC").Find(&assets).Error
	
	// Update current values for all assets
	for i := range assets {
		assets[i].CurrentValue = s.CalculateBookValue(
			assets[i].PurchasePrice,
			assets[i].PurchaseDate,
			assets[i].DepreciationRate,
		)
	}
	
	return assets, err
}

// UpdateAsset updates an existing asset
func (s *AssetService) UpdateAsset(id uint, updates *models.KitchenAsset) error {
	// Check if asset exists
	_, err := s.GetAssetByID(id)
	if err != nil {
		return err
	}

	// Check for duplicate asset code (excluding current asset)
	var existing models.KitchenAsset
	err = s.db.Where("asset_code = ? AND id != ?", updates.AssetCode, id).First(&existing).Error
	if err == nil {
		return ErrDuplicateAssetCode
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Recalculate current value
	updates.CurrentValue = s.CalculateBookValue(updates.PurchasePrice, updates.PurchaseDate, updates.DepreciationRate)

	// Update asset
	return s.db.Model(&models.KitchenAsset{}).Where("id = ?", id).Updates(map[string]interface{}{
		"asset_code":        updates.AssetCode,
		"name":              updates.Name,
		"category":          updates.Category,
		"purchase_date":     updates.PurchaseDate,
		"purchase_price":    updates.PurchasePrice,
		"current_value":     updates.CurrentValue,
		"depreciation_rate": updates.DepreciationRate,
		"condition":         updates.Condition,
		"location":          updates.Location,
		"updated_at":        time.Now(),
	}).Error
}

// DeleteAsset deletes an asset
func (s *AssetService) DeleteAsset(id uint) error {
	result := s.db.Delete(&models.KitchenAsset{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAssetNotFound
	}
	return nil
}

// CalculateBookValue calculates the current book value of an asset
// Formula: Book Value = Purchase Price - (Purchase Price * Depreciation Rate * Years Elapsed)
func (s *AssetService) CalculateBookValue(purchasePrice float64, purchaseDate time.Time, depreciationRate float64) float64 {
	now := time.Now()
	
	// Calculate years elapsed (including fractional years)
	yearsElapsed := now.Sub(purchaseDate).Hours() / (24 * 365.25)
	
	// Calculate accumulated depreciation
	accumulatedDepreciation := purchasePrice * (depreciationRate / 100) * yearsElapsed
	
	// Book value cannot be negative
	bookValue := purchasePrice - accumulatedDepreciation
	if bookValue < 0 {
		bookValue = 0
	}
	
	return bookValue
}

// CalculateAccumulatedDepreciation calculates total depreciation for an asset
func (s *AssetService) CalculateAccumulatedDepreciation(purchasePrice float64, purchaseDate time.Time, depreciationRate float64) float64 {
	now := time.Now()
	yearsElapsed := now.Sub(purchaseDate).Hours() / (24 * 365.25)
	
	accumulatedDepreciation := purchasePrice * (depreciationRate / 100) * yearsElapsed
	
	// Cannot exceed purchase price
	if accumulatedDepreciation > purchasePrice {
		accumulatedDepreciation = purchasePrice
	}
	
	return accumulatedDepreciation
}

// AddMaintenanceRecord adds a maintenance record for an asset
func (s *AssetService) AddMaintenanceRecord(assetID uint, maintenance *models.AssetMaintenance) error {
	// Check if asset exists
	_, err := s.GetAssetByID(assetID)
	if err != nil {
		return err
	}

	maintenance.AssetID = assetID
	return s.db.Create(maintenance).Error
}

// GetMaintenanceRecords retrieves all maintenance records for an asset
func (s *AssetService) GetMaintenanceRecords(assetID uint) ([]models.AssetMaintenance, error) {
	var records []models.AssetMaintenance
	err := s.db.Where("asset_id = ?", assetID).Order("maintenance_date DESC").Find(&records).Error
	return records, err
}

// AssetReport represents asset summary report
type AssetReport struct {
	TotalAssets            int                      `json:"total_assets"`
	TotalPurchaseValue     float64                  `json:"total_purchase_value"`
	TotalCurrentValue      float64                  `json:"total_current_value"`
	TotalDepreciation      float64                  `json:"total_depreciation"`
	AssetsByCategory       map[string]CategoryStats `json:"assets_by_category"`
	AssetsByCondition      map[string]int           `json:"assets_by_condition"`
	TotalMaintenanceCost   float64                  `json:"total_maintenance_cost"`
}

// CategoryStats represents statistics for an asset category
type CategoryStats struct {
	Count         int     `json:"count"`
	PurchaseValue float64 `json:"purchase_value"`
	CurrentValue  float64 `json:"current_value"`
	Depreciation  float64 `json:"depreciation"`
}

// GenerateAssetReport generates a comprehensive asset report
func (s *AssetService) GenerateAssetReport() (*AssetReport, error) {
	var assets []models.KitchenAsset
	err := s.db.Find(&assets).Error
	if err != nil {
		return nil, err
	}

	report := &AssetReport{
		AssetsByCategory:  make(map[string]CategoryStats),
		AssetsByCondition: make(map[string]int),
	}

	for _, asset := range assets {
		// Update current value
		currentValue := s.CalculateBookValue(asset.PurchasePrice, asset.PurchaseDate, asset.DepreciationRate)
		depreciation := asset.PurchasePrice - currentValue

		// Update totals
		report.TotalAssets++
		report.TotalPurchaseValue += asset.PurchasePrice
		report.TotalCurrentValue += currentValue
		report.TotalDepreciation += depreciation

		// Update category stats
		categoryStats := report.AssetsByCategory[asset.Category]
		categoryStats.Count++
		categoryStats.PurchaseValue += asset.PurchasePrice
		categoryStats.CurrentValue += currentValue
		categoryStats.Depreciation += depreciation
		report.AssetsByCategory[asset.Category] = categoryStats

		// Update condition stats
		report.AssetsByCondition[asset.Condition]++
	}

	// Calculate total maintenance cost
	var totalMaintenanceCost float64
	s.db.Model(&models.AssetMaintenance{}).Select("COALESCE(SUM(cost), 0)").Scan(&totalMaintenanceCost)
	report.TotalMaintenanceCost = totalMaintenanceCost

	return report, nil
}

// SearchAssets searches assets by name, category, or condition
func (s *AssetService) SearchAssets(query string, category string, condition string) ([]models.KitchenAsset, error) {
	var assets []models.KitchenAsset
	db := s.db.Model(&models.KitchenAsset{})

	if query != "" {
		db = db.Where("name LIKE ? OR asset_code LIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if category != "" {
		db = db.Where("category = ?", category)
	}

	if condition != "" {
		db = db.Where("condition = ?", condition)
	}

	err := db.Order("name ASC").Find(&assets).Error
	
	// Update current values
	for i := range assets {
		assets[i].CurrentValue = s.CalculateBookValue(
			assets[i].PurchasePrice,
			assets[i].PurchaseDate,
			assets[i].DepreciationRate,
		)
	}
	
	return assets, err
}

// GetDepreciationSchedule generates a depreciation schedule for an asset
func (s *AssetService) GetDepreciationSchedule(assetID uint, years int) ([]DepreciationEntry, error) {
	asset, err := s.GetAssetByID(assetID)
	if err != nil {
		return nil, err
	}

	if years <= 0 {
		years = 5 // Default to 5 years
	}

	schedule := make([]DepreciationEntry, 0, years)
	
	for year := 0; year <= years; year++ {
		futureDate := asset.PurchaseDate.AddDate(year, 0, 0)
		bookValue := s.CalculateBookValue(asset.PurchasePrice, futureDate, asset.DepreciationRate)
		depreciation := asset.PurchasePrice - bookValue
		
		entry := DepreciationEntry{
			Year:                    year,
			Date:                    futureDate,
			BookValue:               bookValue,
			AccumulatedDepreciation: depreciation,
			AnnualDepreciation:      asset.PurchasePrice * (asset.DepreciationRate / 100),
		}
		
		schedule = append(schedule, entry)
	}

	return schedule, nil
}

// DepreciationEntry represents a single entry in depreciation schedule
type DepreciationEntry struct {
	Year                    int       `json:"year"`
	Date                    time.Time `json:"date"`
	BookValue               float64   `json:"book_value"`
	AccumulatedDepreciation float64   `json:"accumulated_depreciation"`
	AnnualDepreciation      float64   `json:"annual_depreciation"`
}
