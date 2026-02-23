package services

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"gorm.io/gorm"
)

// DashboardService handles executive dashboard operations
type DashboardService struct {
	db                     *gorm.DB
	firebaseApp            *firebase.App
	dbClient               *db.Client
	kdsService             *KDSService
	deliveryTaskService    *DeliveryTaskService
	inventoryService       *InventoryService
	cashFlowService        *CashFlowService
	financialReportService *FinancialReportService
	supplierService        *SupplierService
}

// NewDashboardService creates a new dashboard service instance
func NewDashboardService(database *gorm.DB, firebaseApp *firebase.App) (*DashboardService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan Firebase database client: %w", err)
	}

	return &DashboardService{
		db:                     database,
		firebaseApp:            firebaseApp,
		dbClient:               dbClient,
		kdsService:             nil, // Will be initialized when needed
		deliveryTaskService:    NewDeliveryTaskService(database),
		inventoryService:       NewInventoryService(database),
		cashFlowService:        NewCashFlowService(database),
		financialReportService: NewFinancialReportService(database),
		supplierService:        NewSupplierService(database),
	}, nil
}

// KepalaSSPGDashboard represents operational dashboard for Kepala SPPG
type KepalaSSPGDashboard struct {
	ProductionStatus  *ProductionStatus  `json:"production_status"`
	DeliveryStatus    *DeliveryStatus    `json:"delivery_status"`
	CriticalStock     []CriticalStockItem `json:"critical_stock"`
	TodayKPIs         *TodayKPIs         `json:"today_kpis"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// ProductionStatus represents production milestones
type ProductionStatus struct {
	TotalRecipes      int     `json:"total_recipes"`
	RecipesPending    int     `json:"recipes_pending"`
	RecipesCooking    int     `json:"recipes_cooking"`
	RecipesReady      int     `json:"recipes_ready"`
	PackingPending    int     `json:"packing_pending"`
	PackingInProgress int     `json:"packing_in_progress"`
	PackingReady      int     `json:"packing_ready"`
	CompletionRate    float64 `json:"completion_rate"`
}

// DeliveryStatus represents delivery progress
type DeliveryStatus struct {
	TotalDeliveries     int     `json:"total_deliveries"`
	DeliveriesPending   int     `json:"deliveries_pending"`
	DeliveriesInProgress int    `json:"deliveries_in_progress"`
	DeliveriesCompleted int     `json:"deliveries_completed"`
	CompletionRate      float64 `json:"completion_rate"`
}

// CriticalStockItem represents low stock item
type CriticalStockItem struct {
	IngredientID   uint    `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	CurrentStock   float64 `json:"current_stock"`
	MinThreshold   float64 `json:"min_threshold"`
	Unit           string  `json:"unit"`
	DaysRemaining  float64 `json:"days_remaining"`
}

// TodayKPIs represents key performance indicators for today
type TodayKPIs struct {
	PortionsPrepared      int     `json:"portions_prepared"`
	DeliveryRate          float64 `json:"delivery_rate"`
	StockAvailability     float64 `json:"stock_availability"`
	OnTimeDeliveryRate    float64 `json:"on_time_delivery_rate"`
}

// KepalaYayasanDashboard represents strategic dashboard for Kepala Yayasan
type KepalaYayasanDashboard struct {
	BudgetAbsorption      *BudgetAbsorption      `json:"budget_absorption"`
	NutritionDistribution *NutritionDistribution `json:"nutrition_distribution"`
	SupplierPerformance   *SupplierMetrics       `json:"supplier_performance"`
	MonthlyTrend          []MonthlyMetrics       `json:"monthly_trend"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// BudgetAbsorption represents budget usage
type BudgetAbsorption struct {
	TotalBudget       float64                    `json:"total_budget"`
	TotalSpent        float64                    `json:"total_spent"`
	AbsorptionRate    float64                    `json:"absorption_rate"`
	CategoryBreakdown []BudgetCategoryBreakdown  `json:"category_breakdown"`
}

// BudgetCategoryBreakdown represents budget by category
type BudgetCategoryBreakdown struct {
	Category       string  `json:"category"`
	Budget         float64 `json:"budget"`
	Spent          float64 `json:"spent"`
	AbsorptionRate float64 `json:"absorption_rate"`
}

// NutritionDistribution represents distribution metrics
type NutritionDistribution struct {
	TotalPortionsDistributed int     `json:"total_portions_distributed"`
	SchoolsServed            int     `json:"schools_served"`
	StudentsReached          int     `json:"students_reached"`
	AveragePortionsPerSchool float64 `json:"average_portions_per_school"`
}

// SupplierMetrics represents supplier metrics for dashboard
type SupplierMetrics struct {
	TotalSuppliers     int     `json:"total_suppliers"`
	ActiveSuppliers    int     `json:"active_suppliers"`
	AvgOnTimeDelivery  float64 `json:"avg_on_time_delivery"`
	AvgQualityRating   float64 `json:"avg_quality_rating"`
}

// MonthlyMetrics represents monthly trend data
type MonthlyMetrics struct {
	Month              string  `json:"month"`
	Year               int     `json:"year"`
	PortionsDistributed int    `json:"portions_distributed"`
	BudgetSpent        float64 `json:"budget_spent"`
	SchoolsServed      int     `json:"schools_served"`
}

// GetKepalaSSPGDashboard retrieves operational dashboard data
func (s *DashboardService) GetKepalaSSPGDashboard(ctx context.Context) (*KepalaSSPGDashboard, error) {
	dashboard := &KepalaSSPGDashboard{
		UpdatedAt: time.Now(),
	}

	// Get production status
	productionStatus, err := s.getProductionStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan status produksi: %w", err)
	}
	dashboard.ProductionStatus = productionStatus

	// Get delivery status
	deliveryStatus, err := s.getDeliveryStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan status pengiriman: %w", err)
	}
	dashboard.DeliveryStatus = deliveryStatus

	// Get critical stock items
	criticalStock, err := s.getCriticalStock(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan stok kritis: %w", err)
	}
	dashboard.CriticalStock = criticalStock

	// Calculate today's KPIs
	todayKPIs, err := s.calculateTodayKPIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal menghitung KPI hari ini: %w", err)
	}
	dashboard.TodayKPIs = todayKPIs

	return dashboard, nil
}

// getProductionStatus retrieves production status for today
func (s *DashboardService) getProductionStatus(ctx context.Context) (*ProductionStatus, error) {
	today := time.Now().Truncate(24 * time.Hour)
	
	// Get menu items for today
	var totalRecipes int64
	err := s.db.WithContext(ctx).
		Table("menu_items").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("DATE(menu_items.date) = DATE(?)", today).
		Count(&totalRecipes).Error
	if err != nil {
		return nil, err
	}

	// Get recipe statuses from Firebase
	firebasePath := fmt.Sprintf("/kds/cooking/%s", today.Format("2006-01-02"))
	var recipeStatuses map[string]interface{}
	err = s.dbClient.NewRef(firebasePath).Get(ctx, &recipeStatuses)
	if err != nil && err.Error() != "client: no data at ref" {
		return nil, err
	}

	// Count recipes by status
	var pending, cooking, ready int
	if recipeStatuses != nil {
		for _, v := range recipeStatuses {
			if recipeData, ok := v.(map[string]interface{}); ok {
				status, _ := recipeData["status"].(string)
				switch status {
				case "pending":
					pending++
				case "cooking":
					cooking++
				case "ready":
					ready++
				}
			}
		}
	} else {
		// If no Firebase data, all recipes are pending
		pending = int(totalRecipes)
	}

	// Get packing status
	packingPath := fmt.Sprintf("/kds/packing/%s", today.Format("2006-01-02"))
	var packingStatuses map[string]interface{}
	err = s.dbClient.NewRef(packingPath).Get(ctx, &packingStatuses)
	if err != nil && err.Error() != "client: no data at ref" {
		return nil, err
	}

	var packingPending, packingInProgress, packingReady int
	if packingStatuses != nil {
		for _, v := range packingStatuses {
			if packingData, ok := v.(map[string]interface{}); ok {
				status, _ := packingData["status"].(string)
				switch status {
				case "pending":
					packingPending++
				case "packing":
					packingInProgress++
				case "ready":
					packingReady++
				}
			}
		}
	}

	// Calculate completion rate
	completionRate := 0.0
	if totalRecipes > 0 {
		completionRate = (float64(ready) / float64(totalRecipes)) * 100
	}

	return &ProductionStatus{
		TotalRecipes:      int(totalRecipes),
		RecipesPending:    pending,
		RecipesCooking:    cooking,
		RecipesReady:      ready,
		PackingPending:    packingPending,
		PackingInProgress: packingInProgress,
		PackingReady:      packingReady,
		CompletionRate:    completionRate,
	}, nil
}

// getDeliveryStatus retrieves delivery status for today
func (s *DashboardService) getDeliveryStatus(ctx context.Context) (*DeliveryStatus, error) {
	today := time.Now()
	tasks, err := s.deliveryTaskService.GetAllDeliveryTasks(nil, "", &today)
	if err != nil {
		return nil, err
	}

	var pending, inProgress, completed int
	for _, task := range tasks {
		switch task.Status {
		case "pending":
			pending++
		case "in_progress":
			inProgress++
		case "completed":
			completed++
		}
	}

	total := len(tasks)
	completionRate := 0.0
	if total > 0 {
		completionRate = (float64(completed) / float64(total)) * 100
	}

	return &DeliveryStatus{
		TotalDeliveries:      total,
		DeliveriesPending:    pending,
		DeliveriesInProgress: inProgress,
		DeliveriesCompleted:  completed,
		CompletionRate:       completionRate,
	}, nil
}

// getCriticalStock retrieves items below minimum threshold
func (s *DashboardService) getCriticalStock(ctx context.Context) ([]CriticalStockItem, error) {
	alerts, err := s.inventoryService.CheckLowStock()
	if err != nil {
		return nil, err
	}

	criticalItems := make([]CriticalStockItem, len(alerts))
	for i, alert := range alerts {
		criticalItems[i] = CriticalStockItem{
			IngredientID:   alert.IngredientID,
			IngredientName: alert.IngredientName,
			CurrentStock:   alert.CurrentStock,
			MinThreshold:   alert.MinThreshold,
			Unit:           alert.Unit,
			DaysRemaining:  alert.DaysRemaining,
		}
	}

	return criticalItems, nil
}

// calculateTodayKPIs calculates key performance indicators for today
func (s *DashboardService) calculateTodayKPIs(ctx context.Context) (*TodayKPIs, error) {
	today := time.Now()
	
	// Calculate portions prepared (from menu items)
	var portionsPrepared int64
	err := s.db.WithContext(ctx).
		Table("menu_items").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("DATE(menu_items.date) = DATE(?)", today).
		Select("COALESCE(SUM(portions), 0)").
		Scan(&portionsPrepared).Error
	if err != nil {
		return nil, err
	}

	// Calculate delivery rate
	tasks, err := s.deliveryTaskService.GetAllDeliveryTasks(nil, "", &today)
	if err != nil {
		return nil, err
	}

	deliveryRate := 0.0
	if len(tasks) > 0 {
		completed := 0
		for _, task := range tasks {
			if task.Status == "completed" {
				completed++
			}
		}
		deliveryRate = (float64(completed) / float64(len(tasks))) * 100
	}

	// Calculate stock availability (percentage of items above threshold)
	allInventory, err := s.inventoryService.GetAllInventory()
	if err != nil {
		return nil, err
	}

	stockAvailability := 0.0
	if len(allInventory) > 0 {
		aboveThreshold := 0
		for _, item := range allInventory {
			if item.Quantity >= item.MinThreshold {
				aboveThreshold++
			}
		}
		stockAvailability = (float64(aboveThreshold) / float64(len(allInventory))) * 100
	}

	// Calculate on-time delivery rate (placeholder - would need delivery time tracking)
	onTimeDeliveryRate := 95.0 // Placeholder

	return &TodayKPIs{
		PortionsPrepared:   int(portionsPrepared),
		DeliveryRate:       deliveryRate,
		StockAvailability:  stockAvailability,
		OnTimeDeliveryRate: onTimeDeliveryRate,
	}, nil
}

// GetKepalaYayasanDashboard retrieves strategic dashboard data
func (s *DashboardService) GetKepalaYayasanDashboard(ctx context.Context, startDate, endDate time.Time) (*KepalaYayasanDashboard, error) {
	dashboard := &KepalaYayasanDashboard{
		UpdatedAt: time.Now(),
	}

	// Get budget absorption
	budgetAbsorption, err := s.getBudgetAbsorption(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan penyerapan anggaran: %w", err)
	}
	dashboard.BudgetAbsorption = budgetAbsorption

	// Get nutrition distribution
	nutritionDistribution, err := s.getNutritionDistribution(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan distribusi nutrisi: %w", err)
	}
	dashboard.NutritionDistribution = nutritionDistribution

	// Get supplier performance
	supplierPerformance, err := s.getSupplierPerformance(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan performa supplier: %w", err)
	}
	dashboard.SupplierPerformance = supplierPerformance

	// Get monthly trend
	monthlyTrend, err := s.getMonthlyTrend(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan tren bulanan: %w", err)
	}
	dashboard.MonthlyTrend = monthlyTrend

	return dashboard, nil
}

// getBudgetAbsorption calculates budget absorption for the period
func (s *DashboardService) getBudgetAbsorption(ctx context.Context, startDate, endDate time.Time) (*BudgetAbsorption, error) {
	// Get budget targets for the period
	var budgetTargets []struct {
		Category string
		Target   float64
	}
	err := s.db.WithContext(ctx).
		Table("budget_targets").
		Select("category, SUM(target) as target").
		Where("year = ? AND month >= ? AND month <= ?",
			startDate.Year(),
			int(startDate.Month()),
			int(endDate.Month())).
		Group("category").
		Scan(&budgetTargets).Error
	if err != nil {
		return nil, err
	}

	// Get actual spending by category
	var actualSpending []struct {
		Category string
		Amount   float64
	}
	err = s.db.WithContext(ctx).
		Table("cash_flow_entries").
		Select("category, SUM(amount) as amount").
		Where("type = ? AND date BETWEEN ? AND ?", "expense", startDate, endDate).
		Group("category").
		Scan(&actualSpending).Error
	if err != nil {
		return nil, err
	}

	// Build category breakdown
	budgetMap := make(map[string]float64)
	for _, bt := range budgetTargets {
		budgetMap[bt.Category] = bt.Target
	}

	actualMap := make(map[string]float64)
	for _, as := range actualSpending {
		actualMap[as.Category] = as.Amount
	}

	var categoryBreakdown []BudgetCategoryBreakdown
	var totalBudget, totalSpent float64

	// Combine all categories
	allCategories := make(map[string]bool)
	for cat := range budgetMap {
		allCategories[cat] = true
	}
	for cat := range actualMap {
		allCategories[cat] = true
	}

	for category := range allCategories {
		budget := budgetMap[category]
		spent := actualMap[category]
		absorptionRate := 0.0
		if budget > 0 {
			absorptionRate = (spent / budget) * 100
		}

		categoryBreakdown = append(categoryBreakdown, BudgetCategoryBreakdown{
			Category:       category,
			Budget:         budget,
			Spent:          spent,
			AbsorptionRate: absorptionRate,
		})

		totalBudget += budget
		totalSpent += spent
	}

	overallAbsorptionRate := 0.0
	if totalBudget > 0 {
		overallAbsorptionRate = (totalSpent / totalBudget) * 100
	}

	return &BudgetAbsorption{
		TotalBudget:       totalBudget,
		TotalSpent:        totalSpent,
		AbsorptionRate:    overallAbsorptionRate,
		CategoryBreakdown: categoryBreakdown,
	}, nil
}

// getNutritionDistribution calculates nutrition distribution metrics
func (s *DashboardService) getNutritionDistribution(ctx context.Context, startDate, endDate time.Time) (*NutritionDistribution, error) {
	// Get total portions distributed
	var totalPortions int64
	err := s.db.WithContext(ctx).
		Table("delivery_tasks").
		Where("status = ? AND task_date BETWEEN ? AND ?", "completed", startDate, endDate).
		Select("COALESCE(SUM(portions), 0)").
		Scan(&totalPortions).Error
	if err != nil {
		return nil, err
	}

	// Get schools served
	var schoolsServed int64
	err = s.db.WithContext(ctx).
		Table("delivery_tasks").
		Where("status = ? AND task_date BETWEEN ? AND ?", "completed", startDate, endDate).
		Distinct("school_id").
		Count(&schoolsServed).Error
	if err != nil {
		return nil, err
	}

	// Get total students reached
	var studentsReached int64
	err = s.db.WithContext(ctx).
		Table("schools").
		Joins("JOIN delivery_tasks ON schools.id = delivery_tasks.school_id").
		Where("delivery_tasks.status = ? AND delivery_tasks.task_date BETWEEN ? AND ?", "completed", startDate, endDate).
		Select("COALESCE(SUM(DISTINCT schools.student_count), 0)").
		Scan(&studentsReached).Error
	if err != nil {
		return nil, err
	}

	// Calculate average portions per school
	avgPortionsPerSchool := 0.0
	if schoolsServed > 0 {
		avgPortionsPerSchool = float64(totalPortions) / float64(schoolsServed)
	}

	return &NutritionDistribution{
		TotalPortionsDistributed: int(totalPortions),
		SchoolsServed:            int(schoolsServed),
		StudentsReached:          int(studentsReached),
		AveragePortionsPerSchool: avgPortionsPerSchool,
	}, nil
}

// getSupplierPerformance calculates supplier performance metrics
func (s *DashboardService) getSupplierPerformance(ctx context.Context) (*SupplierMetrics, error) {
	// Get total suppliers
	var totalSuppliers int64
	err := s.db.WithContext(ctx).Model(&struct{}{}).Table("suppliers").Count(&totalSuppliers).Error
	if err != nil {
		return nil, err
	}

	// Get active suppliers
	var activeSuppliers int64
	err = s.db.WithContext(ctx).Model(&struct{}{}).Table("suppliers").Where("is_active = ?", true).Count(&activeSuppliers).Error
	if err != nil {
		return nil, err
	}

	// Calculate average on-time delivery and quality rating
	var avgMetrics struct {
		AvgOnTimeDelivery float64
		AvgQualityRating  float64
	}
	err = s.db.WithContext(ctx).
		Table("suppliers").
		Where("is_active = ?", true).
		Select("COALESCE(AVG(on_time_delivery), 0) as avg_on_time_delivery, COALESCE(AVG(quality_rating), 0) as avg_quality_rating").
		Scan(&avgMetrics).Error
	if err != nil {
		return nil, err
	}

	return &SupplierMetrics{
		TotalSuppliers:    int(totalSuppliers),
		ActiveSuppliers:   int(activeSuppliers),
		AvgOnTimeDelivery: avgMetrics.AvgOnTimeDelivery,
		AvgQualityRating:  avgMetrics.AvgQualityRating,
	}, nil
}

// getMonthlyTrend calculates monthly trend data
func (s *DashboardService) getMonthlyTrend(ctx context.Context, startDate, endDate time.Time) ([]MonthlyMetrics, error) {
	var trend []MonthlyMetrics

	// Iterate through each month in the date range
	currentDate := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !currentDate.After(endMonth) {
		monthStart := currentDate
		monthEnd := monthStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		// Get portions distributed
		var portionsDistributed int64
		s.db.WithContext(ctx).
			Table("delivery_tasks").
			Where("status = ? AND task_date BETWEEN ? AND ?", "completed", monthStart, monthEnd).
			Select("COALESCE(SUM(portions), 0)").
			Scan(&portionsDistributed)

		// Get budget spent
		var budgetSpent float64
		s.db.WithContext(ctx).
			Table("cash_flow_entries").
			Where("type = ? AND date BETWEEN ? AND ?", "expense", monthStart, monthEnd).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&budgetSpent)

		// Get schools served
		var schoolsServed int64
		s.db.WithContext(ctx).
			Table("delivery_tasks").
			Where("status = ? AND task_date BETWEEN ? AND ?", "completed", monthStart, monthEnd).
			Distinct("school_id").
			Count(&schoolsServed)

		trend = append(trend, MonthlyMetrics{
			Month:               currentDate.Format("January"),
			Year:                currentDate.Year(),
			PortionsDistributed: int(portionsDistributed),
			BudgetSpent:         budgetSpent,
			SchoolsServed:       int(schoolsServed),
		})

		// Move to next month
		currentDate = currentDate.AddDate(0, 1, 0)
	}

	return trend, nil
}

// SyncKepalaSSPGDashboardToFirebase syncs Kepala SPPG dashboard to Firebase
func (s *DashboardService) SyncKepalaSSPGDashboardToFirebase(ctx context.Context) error {
	dashboard, err := s.GetKepalaSSPGDashboard(ctx)
	if err != nil {
		return err
	}

	firebasePath := "/dashboard/kepala_sppg"
	err = s.dbClient.NewRef(firebasePath).Set(ctx, dashboard)
	if err != nil {
		return fmt.Errorf("gagal sync dashboard ke Firebase: %w", err)
	}

	return nil
}

// SyncKepalaYayasanDashboardToFirebase syncs Kepala Yayasan dashboard to Firebase
func (s *DashboardService) SyncKepalaYayasanDashboardToFirebase(ctx context.Context, startDate, endDate time.Time) error {
	dashboard, err := s.GetKepalaYayasanDashboard(ctx, startDate, endDate)
	if err != nil {
		return err
	}

	firebasePath := "/dashboard/kepala_yayasan"
	err = s.dbClient.NewRef(firebasePath).Set(ctx, dashboard)
	if err != nil {
		return fmt.Errorf("gagal sync dashboard ke Firebase: %w", err)
	}

	return nil
}

// ExportDashboardData exports dashboard data for reporting
func (s *DashboardService) ExportDashboardData(ctx context.Context, dashboardType string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var data map[string]interface{}

	switch dashboardType {
	case "kepala_sppg":
		dashboard, err := s.GetKepalaSSPGDashboard(ctx)
		if err != nil {
			return nil, err
		}
		data = map[string]interface{}{
			"type":      "Kepala SPPG Dashboard",
			"dashboard": dashboard,
		}

	case "kepala_yayasan":
		dashboard, err := s.GetKepalaYayasanDashboard(ctx, startDate, endDate)
		if err != nil {
			return nil, err
		}
		data = map[string]interface{}{
			"type":       "Kepala Yayasan Dashboard",
			"dashboard":  dashboard,
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		}

	default:
		return nil, fmt.Errorf("tipe dashboard tidak valid: %s", dashboardType)
	}

	return data, nil
}
