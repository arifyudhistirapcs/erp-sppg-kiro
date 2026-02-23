package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrInvalidDateRange = errors.New("rentang tanggal tidak valid")
	ErrReportGeneration = errors.New("gagal membuat laporan")
)

// FinancialReportService handles financial reporting business logic
type FinancialReportService struct {
	db              *gorm.DB
	cashFlowService *CashFlowService
	assetService    *AssetService
}

// NewFinancialReportService creates a new financial report service
func NewFinancialReportService(db *gorm.DB) *FinancialReportService {
	return &FinancialReportService{
		db:              db,
		cashFlowService: NewCashFlowService(db),
		assetService:    NewAssetService(db),
	}
}

// FinancialReport represents a comprehensive financial report
type FinancialReport struct {
	ReportPeriod      string                    `json:"report_period"`
	StartDate         time.Time                 `json:"start_date"`
	EndDate           time.Time                 `json:"end_date"`
	GeneratedAt       time.Time                 `json:"generated_at"`
	CashFlowSummary   *CashFlowSummary          `json:"cash_flow_summary"`
	BudgetComparison  *BudgetComparison         `json:"budget_comparison,omitempty"`
	AssetSummary      *AssetSummary             `json:"asset_summary,omitempty"`
	CategoryBreakdown []CategoryBreakdown       `json:"category_breakdown"`
	MonthlyTrend      []MonthlyData             `json:"monthly_trend,omitempty"`
}

// BudgetComparison represents budget vs actual comparison
type BudgetComparison struct {
	Categories []BudgetCategoryComparison `json:"categories"`
	TotalBudget float64                   `json:"total_budget"`
	TotalActual float64                   `json:"total_actual"`
	Variance    float64                   `json:"variance"`
	VariancePercent float64               `json:"variance_percent"`
}

// BudgetCategoryComparison represents budget comparison for a category
type BudgetCategoryComparison struct {
	Category        string  `json:"category"`
	Budget          float64 `json:"budget"`
	Actual          float64 `json:"actual"`
	Variance        float64 `json:"variance"`
	VariancePercent float64 `json:"variance_percent"`
}

// AssetSummary represents asset summary in financial report
type AssetSummary struct {
	TotalAssets       int     `json:"total_assets"`
	TotalPurchaseValue float64 `json:"total_purchase_value"`
	TotalCurrentValue  float64 `json:"total_current_value"`
	TotalDepreciation  float64 `json:"total_depreciation"`
}

// CategoryBreakdown represents expense breakdown by category
type CategoryBreakdown struct {
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
	Count       int     `json:"count"`
}

// MonthlyData represents monthly financial data
type MonthlyData struct {
	Month       string  `json:"month"`
	Year        int     `json:"year"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	NetCashFlow float64 `json:"net_cash_flow"`
}

// GenerateFinancialReport generates a comprehensive financial report
func (s *FinancialReportService) GenerateFinancialReport(startDate, endDate time.Time, includeBudget, includeAssets, includeTrend bool) (*FinancialReport, error) {
	// Validate date range
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	report := &FinancialReport{
		StartDate:   startDate,
		EndDate:     endDate,
		GeneratedAt: time.Now(),
	}

	// Set report period description
	report.ReportPeriod = s.formatReportPeriod(startDate, endDate)

	// Get cash flow summary
	cashFlowSummary, err := s.cashFlowService.GetCashFlowSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.CashFlowSummary = cashFlowSummary

	// Generate category breakdown
	categoryBreakdown, err := s.generateCategoryBreakdown(startDate, endDate)
	if err != nil {
		return nil, err
	}
	report.CategoryBreakdown = categoryBreakdown

	// Include budget comparison if requested
	if includeBudget {
		budgetComparison, err := s.generateBudgetComparison(startDate, endDate)
		if err == nil {
			report.BudgetComparison = budgetComparison
		}
	}

	// Include asset summary if requested
	if includeAssets {
		assetSummary, err := s.generateAssetSummary()
		if err == nil {
			report.AssetSummary = assetSummary
		}
	}

	// Include monthly trend if requested
	if includeTrend {
		monthlyTrend, err := s.generateMonthlyTrend(startDate, endDate)
		if err == nil {
			report.MonthlyTrend = monthlyTrend
		}
	}

	return report, nil
}

// generateCategoryBreakdown generates expense breakdown by category
func (s *FinancialReportService) generateCategoryBreakdown(startDate, endDate time.Time) ([]CategoryBreakdown, error) {
	var results []struct {
		Category string
		Amount   float64
		Count    int64
	}

	err := s.db.Model(&models.CashFlowEntry{}).
		Select("category, SUM(amount) as amount, COUNT(*) as count").
		Where("type = ? AND date BETWEEN ? AND ?", "expense", startDate, endDate).
		Group("category").
		Order("amount DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Calculate total for percentage
	var total float64
	for _, r := range results {
		total += r.Amount
	}

	breakdown := make([]CategoryBreakdown, len(results))
	for i, r := range results {
		percentage := 0.0
		if total > 0 {
			percentage = (r.Amount / total) * 100
		}

		breakdown[i] = CategoryBreakdown{
			Category:   r.Category,
			Amount:     r.Amount,
			Percentage: percentage,
			Count:      int(r.Count),
		}
	}

	return breakdown, nil
}

// generateBudgetComparison generates budget vs actual comparison
func (s *FinancialReportService) generateBudgetComparison(startDate, endDate time.Time) (*BudgetComparison, error) {
	comparison := &BudgetComparison{
		Categories: make([]BudgetCategoryComparison, 0),
	}

	// Get budget targets for the period
	var budgetTargets []models.BudgetTarget
	err := s.db.Where("year = ? AND month >= ? AND month <= ?",
		startDate.Year(),
		int(startDate.Month()),
		int(endDate.Month())).
		Find(&budgetTargets).Error

	if err != nil {
		return nil, err
	}

	// Aggregate budget by category
	budgetByCategory := make(map[string]float64)
	for _, target := range budgetTargets {
		budgetByCategory[target.Category] += target.Target
	}

	// Get actual expenses by category
	actualByCategory := make(map[string]float64)
	var cashFlowEntries []models.CashFlowEntry
	err = s.db.Where("type = ? AND date BETWEEN ? AND ?", "expense", startDate, endDate).
		Find(&cashFlowEntries).Error
	if err != nil {
		return nil, err
	}

	for _, entry := range cashFlowEntries {
		actualByCategory[entry.Category] += entry.Amount
	}

	// Build comparison for each category
	for category, budget := range budgetByCategory {
		actual := actualByCategory[category]
		variance := budget - actual
		variancePercent := 0.0
		if budget > 0 {
			variancePercent = (variance / budget) * 100
		}

		comparison.Categories = append(comparison.Categories, BudgetCategoryComparison{
			Category:        category,
			Budget:          budget,
			Actual:          actual,
			Variance:        variance,
			VariancePercent: variancePercent,
		})

		comparison.TotalBudget += budget
		comparison.TotalActual += actual
	}

	comparison.Variance = comparison.TotalBudget - comparison.TotalActual
	if comparison.TotalBudget > 0 {
		comparison.VariancePercent = (comparison.Variance / comparison.TotalBudget) * 100
	}

	return comparison, nil
}

// generateAssetSummary generates asset summary for financial report
func (s *FinancialReportService) generateAssetSummary() (*AssetSummary, error) {
	assetReport, err := s.assetService.GenerateAssetReport()
	if err != nil {
		return nil, err
	}

	return &AssetSummary{
		TotalAssets:        assetReport.TotalAssets,
		TotalPurchaseValue: assetReport.TotalPurchaseValue,
		TotalCurrentValue:  assetReport.TotalCurrentValue,
		TotalDepreciation:  assetReport.TotalDepreciation,
	}, nil
}

// generateMonthlyTrend generates monthly cash flow trend
func (s *FinancialReportService) generateMonthlyTrend(startDate, endDate time.Time) ([]MonthlyData, error) {
	trend := make([]MonthlyData, 0)

	// Iterate through each month in the date range
	currentDate := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !currentDate.After(endMonth) {
		monthStart := currentDate
		monthEnd := monthStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		// Get cash flow summary for the month
		summary, err := s.cashFlowService.GetCashFlowSummary(monthStart, monthEnd)
		if err != nil {
			return nil, err
		}

		monthData := MonthlyData{
			Month:       currentDate.Format("January"),
			Year:        currentDate.Year(),
			Income:      summary.TotalIncome,
			Expense:     summary.TotalExpense,
			NetCashFlow: summary.NetCashFlow,
		}

		trend = append(trend, monthData)

		// Move to next month
		currentDate = currentDate.AddDate(0, 1, 0)
	}

	return trend, nil
}

// formatReportPeriod formats the report period description
func (s *FinancialReportService) formatReportPeriod(startDate, endDate time.Time) string {
	if startDate.Year() == endDate.Year() && startDate.Month() == endDate.Month() {
		return fmt.Sprintf("%s %d", startDate.Format("January"), startDate.Year())
	}

	if startDate.Year() == endDate.Year() {
		return fmt.Sprintf("%s - %s %d",
			startDate.Format("January"),
			endDate.Format("January"),
			startDate.Year())
	}

	return fmt.Sprintf("%s %d - %s %d",
		startDate.Format("January"),
		startDate.Year(),
		endDate.Format("January"),
		endDate.Year())
}

// ExportFormat represents export format type
type ExportFormat string

const (
	ExportFormatPDF   ExportFormat = "pdf"
	ExportFormatExcel ExportFormat = "excel"
)

// ExportOptions represents options for exporting reports
type ExportOptions struct {
	Format         ExportFormat
	IncludeBudget  bool
	IncludeAssets  bool
	IncludeTrend   bool
	IncludeCharts  bool
}

// ExportFinancialReport exports a financial report in the specified format
func (s *FinancialReportService) ExportFinancialReport(startDate, endDate time.Time, options ExportOptions) ([]byte, string, error) {
	// Generate the report
	report, err := s.GenerateFinancialReport(
		startDate,
		endDate,
		options.IncludeBudget,
		options.IncludeAssets,
		options.IncludeTrend,
	)
	if err != nil {
		return nil, "", err
	}

	// Export based on format
	switch options.Format {
	case ExportFormatPDF:
		return s.exportToPDF(report, options)
	case ExportFormatExcel:
		return s.exportToExcel(report, options)
	default:
		return nil, "", errors.New("format export tidak didukung")
	}
}

// exportToPDF exports report to PDF format
func (s *FinancialReportService) exportToPDF(report *FinancialReport, options ExportOptions) ([]byte, string, error) {
	// TODO: Implement PDF generation using a library like gofpdf or wkhtmltopdf
	// For now, return a placeholder
	
	// This is a placeholder implementation
	// In production, you would use a PDF library to generate the actual PDF
	filename := fmt.Sprintf("laporan_keuangan_%s.pdf", time.Now().Format("20060102_150405"))
	
	return nil, filename, errors.New("export PDF belum diimplementasikan - memerlukan library PDF")
}

// exportToExcel exports report to Excel format
func (s *FinancialReportService) exportToExcel(report *FinancialReport, options ExportOptions) ([]byte, string, error) {
	// TODO: Implement Excel generation using a library like excelize
	// For now, return a placeholder
	
	// This is a placeholder implementation
	// In production, you would use an Excel library to generate the actual file
	filename := fmt.Sprintf("laporan_keuangan_%s.xlsx", time.Now().Format("20060102_150405"))
	
	return nil, filename, errors.New("export Excel belum diimplementasikan - memerlukan library Excel")
}

// GetDailyCashFlow retrieves daily cash flow summary
func (s *FinancialReportService) GetDailyCashFlow(date time.Time) (*CashFlowSummary, error) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	
	return s.cashFlowService.GetCashFlowSummary(startDate, endDate)
}

// GetWeeklyCashFlow retrieves weekly cash flow summary
func (s *FinancialReportService) GetWeeklyCashFlow(year int, week int) (*CashFlowSummary, error) {
	// Calculate start of week (Monday)
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Find the first Monday of the year
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, 1)
	}
	
	// Add weeks
	startDate = startDate.AddDate(0, 0, (week-1)*7)
	endDate := startDate.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	
	return s.cashFlowService.GetCashFlowSummary(startDate, endDate)
}

// GetQuarterlyCashFlow retrieves quarterly cash flow summary
func (s *FinancialReportService) GetQuarterlyCashFlow(year int, quarter int) (*CashFlowSummary, error) {
	if quarter < 1 || quarter > 4 {
		return nil, errors.New("kuartal harus antara 1 dan 4")
	}
	
	startMonth := (quarter-1)*3 + 1
	startDate := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 3, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	
	return s.cashFlowService.GetCashFlowSummary(startDate, endDate)
}

// GetCustomPeriodReport generates a report for a custom date range
func (s *FinancialReportService) GetCustomPeriodReport(startDate, endDate time.Time) (*FinancialReport, error) {
	return s.GenerateFinancialReport(startDate, endDate, true, true, true)
}
