package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrCashFlowNotFound   = errors.New("cash flow entry tidak ditemukan")
	ErrCashFlowValidation = errors.New("validasi cash flow gagal")
	ErrInvalidCategory    = errors.New("kategori tidak valid")
	ErrInvalidType        = errors.New("tipe transaksi tidak valid")
)

// CashFlowService handles cash flow business logic
type CashFlowService struct {
	db *gorm.DB
}

// NewCashFlowService creates a new cash flow service
func NewCashFlowService(db *gorm.DB) *CashFlowService {
	return &CashFlowService{
		db: db,
	}
}

// CreateCashFlowEntry creates a new cash flow entry
func (s *CashFlowService) CreateCashFlowEntry(entry *models.CashFlowEntry) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.CreateCashFlowEntryWithTx(tx, entry)
	})
}

// CreateCashFlowEntryWithTx creates a new cash flow entry within a transaction
func (s *CashFlowService) CreateCashFlowEntryWithTx(tx *gorm.DB, entry *models.CashFlowEntry) error {
	// Validate type
	if entry.Type != "income" && entry.Type != "expense" {
		return ErrInvalidType
	}

	// Validate category
	validCategories := []string{"bahan_baku", "gaji", "utilitas", "operasional"}
	isValidCategory := false
	for _, cat := range validCategories {
		if entry.Category == cat {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return ErrInvalidCategory
	}

	// Generate transaction ID if not provided
	if entry.TransactionID == "" {
		transactionID, err := s.generateTransactionID()
		if err != nil {
			return err
		}
		entry.TransactionID = transactionID
	}

	// Set date if not provided
	if entry.Date.IsZero() {
		entry.Date = time.Now()
	}

	return tx.Create(entry).Error
}

// GetCashFlowEntryByID retrieves a cash flow entry by ID
func (s *CashFlowService) GetCashFlowEntryByID(id uint) (*models.CashFlowEntry, error) {
	var entry models.CashFlowEntry
	err := s.db.Preload("Creator").First(&entry, id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCashFlowNotFound
		}
		return nil, err
	}

	return &entry, nil
}

// GetAllCashFlowEntries retrieves all cash flow entries with optional filters
func (s *CashFlowService) GetAllCashFlowEntries(category, entryType string, startDate, endDate *time.Time) ([]models.CashFlowEntry, error) {
	query := s.db.Preload("Creator")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if entryType != "" {
		query = query.Where("type = ?", entryType)
	}

	if startDate != nil && endDate != nil {
		query = query.Where("date BETWEEN ? AND ?", *startDate, *endDate)
	}

	var entries []models.CashFlowEntry
	err := query.Order("date DESC").Find(&entries).Error
	return entries, err
}

// CashFlowSummary represents a summary of cash flow
type CashFlowSummary struct {
	TotalIncome      float64            `json:"total_income"`
	TotalExpense     float64            `json:"total_expense"`
	NetCashFlow      float64            `json:"net_cash_flow"`
	ByCategory       map[string]float64 `json:"by_category"`
	StartDate        time.Time          `json:"start_date"`
	EndDate          time.Time          `json:"end_date"`
}

// GetCashFlowSummary generates a cash flow summary for a date range
func (s *CashFlowService) GetCashFlowSummary(startDate, endDate time.Time) (*CashFlowSummary, error) {
	summary := &CashFlowSummary{
		StartDate:  startDate,
		EndDate:    endDate,
		ByCategory: make(map[string]float64),
	}

	// Get all entries in date range
	entries, err := s.GetAllCashFlowEntries("", "", &startDate, &endDate)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	for _, entry := range entries {
		if entry.Type == "income" {
			summary.TotalIncome += entry.Amount
		} else if entry.Type == "expense" {
			summary.TotalExpense += entry.Amount
		}

		// Aggregate by category
		if entry.Type == "expense" {
			summary.ByCategory[entry.Category] += entry.Amount
		}
	}

	summary.NetCashFlow = summary.TotalIncome - summary.TotalExpense

	return summary, nil
}

// GetRunningBalance calculates running balance for a category
func (s *CashFlowService) GetRunningBalance(category string, upToDate time.Time) (float64, error) {
	var totalIncome, totalExpense float64

	// Get total income
	s.db.Model(&models.CashFlowEntry{}).
		Where("category = ? AND type = ? AND date <= ?", category, "income", upToDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome)

	// Get total expense
	s.db.Model(&models.CashFlowEntry{}).
		Where("category = ? AND type = ? AND date <= ?", category, "expense", upToDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	return totalIncome - totalExpense, nil
}

// GetCashFlowByCategory retrieves cash flow entries by category
func (s *CashFlowService) GetCashFlowByCategory(category string, startDate, endDate time.Time) ([]models.CashFlowEntry, error) {
	var entries []models.CashFlowEntry
	err := s.db.Preload("Creator").
		Where("category = ? AND date BETWEEN ? AND ?", category, startDate, endDate).
		Order("date DESC").
		Find(&entries).Error
	return entries, err
}

// GetCashFlowByReference retrieves cash flow entries by reference
func (s *CashFlowService) GetCashFlowByReference(reference string) ([]models.CashFlowEntry, error) {
	var entries []models.CashFlowEntry
	err := s.db.Preload("Creator").
		Where("reference = ?", reference).
		Order("date DESC").
		Find(&entries).Error
	return entries, err
}

// UpdateCashFlowEntry updates an existing cash flow entry
func (s *CashFlowService) UpdateCashFlowEntry(id uint, updates *models.CashFlowEntry) error {
	// Check if entry exists
	_, err := s.GetCashFlowEntryByID(id)
	if err != nil {
		return err
	}

	// Validate type if provided
	if updates.Type != "" && updates.Type != "income" && updates.Type != "expense" {
		return ErrInvalidType
	}

	// Validate category if provided
	if updates.Category != "" {
		validCategories := []string{"bahan_baku", "gaji", "utilitas", "operasional"}
		isValidCategory := false
		for _, cat := range validCategories {
			if updates.Category == cat {
				isValidCategory = true
				break
			}
		}
		if !isValidCategory {
			return ErrInvalidCategory
		}
	}

	// Update entry
	return s.db.Model(&models.CashFlowEntry{}).Where("id = ?", id).Updates(map[string]interface{}{
		"date":        updates.Date,
		"category":    updates.Category,
		"type":        updates.Type,
		"amount":      updates.Amount,
		"description": updates.Description,
		"reference":   updates.Reference,
	}).Error
}

// DeleteCashFlowEntry deletes a cash flow entry
func (s *CashFlowService) DeleteCashFlowEntry(id uint) error {
	result := s.db.Delete(&models.CashFlowEntry{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCashFlowNotFound
	}
	return nil
}

// generateTransactionID generates a unique transaction ID
func (s *CashFlowService) generateTransactionID() (string, error) {
	// Format: TXN-YYYYMMDD-XXXX
	now := time.Now()
	datePrefix := now.Format("20060102")
	
	// Count transactions created today
	var count int64
	s.db.Model(&models.CashFlowEntry{}).
		Where("transaction_id LIKE ?", fmt.Sprintf("TXN-%s-%%", datePrefix)).
		Count(&count)
	
	// Generate transaction ID
	transactionID := fmt.Sprintf("TXN-%s-%04d", datePrefix, count+1)
	
	// Check if it already exists (race condition protection)
	var existing models.CashFlowEntry
	err := s.db.Where("transaction_id = ?", transactionID).First(&existing).Error
	if err == nil {
		// If exists, try with incremented number
		transactionID = fmt.Sprintf("TXN-%s-%04d", datePrefix, count+2)
	}
	
	return transactionID, nil
}

// GetMonthlyCashFlow retrieves monthly cash flow summary
func (s *CashFlowService) GetMonthlyCashFlow(year int, month int) (*CashFlowSummary, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	
	return s.GetCashFlowSummary(startDate, endDate)
}

// GetYearlyCashFlow retrieves yearly cash flow summary
func (s *CashFlowService) GetYearlyCashFlow(year int) (*CashFlowSummary, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	
	return s.GetCashFlowSummary(startDate, endDate)
}
