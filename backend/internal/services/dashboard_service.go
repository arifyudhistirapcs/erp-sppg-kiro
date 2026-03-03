package services

import (
	"context"
	"fmt"
	"log"
	"math"
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
	var dbClient *db.Client
	
	// Try to get Firebase database client, but don't fail if Firebase is not available
	if firebaseApp != nil {
		ctx := context.Background()
		var err error
		dbClient, err = firebaseApp.Database(ctx)
		if err != nil {
			log.Printf("Warning: Failed to get Firebase database client: %v. Using dummy data mode.", err)
			dbClient = nil
		}
	} else {
		log.Println("Warning: Firebase app is nil. Dashboard will use dummy data.")
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

// roundToDecimal rounds a float64 to specified decimal places
func roundToDecimal(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// KepalaSSPGDashboard represents operational dashboard for Kepala SPPG
type KepalaSSPGDashboard struct {
	ProductionStatus  *ProductionStatus   `json:"production_status"`
	DeliveryStatus    *DeliveryStatus     `json:"delivery_status"`
	PickupStatus      *PickupStatus       `json:"pickup_status"`
	CleaningStatus    *CleaningStatus     `json:"cleaning_status"`
	ProductionDetails []SchoolDetail      `json:"production_details"`
	DeliveryDetails   []SchoolDetail      `json:"delivery_details"`
	PickupDetails     []SchoolDetail      `json:"pickup_details"`
	CleaningDetails   []SchoolDetail      `json:"cleaning_details"`
	CriticalStock     []CriticalStockItem `json:"critical_stock"`
	TodayKPIs         *TodayKPIs          `json:"today_kpis"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// SchoolDetail represents school-level detail for dashboard tables
type SchoolDetail struct {
	SchoolID   uint   `json:"school_id"`
	SchoolName string `json:"school_name"`
	Portions   int    `json:"portions"`
	Status     string `json:"status"`
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

// DeliveryStatus represents delivery progress (stages 6-9: delivery to school)
type DeliveryStatus struct {
	TotalDeliveries      int            `json:"total_deliveries"`
	StatusBreakdown      []StatusCount  `json:"status_breakdown"`
	CompletionRate       float64        `json:"completion_rate"`
}

// StatusCount represents count per status for charts
type StatusCount struct {
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
	Count       int    `json:"count"`
}

// PickupStatus represents ompreng pickup progress (stages 10-13: pickup from school back to SPPG)
type PickupStatus struct {
	TotalPickups    int           `json:"total_pickups"`
	StatusBreakdown []StatusCount `json:"status_breakdown"`
	CompletionRate  float64       `json:"completion_rate"`
}

// CleaningStatus represents cleaning progress
type CleaningStatus struct {
	TotalItems      int     `json:"total_items"`
	ItemsPending    int     `json:"items_pending"`
	ItemsInProgress int     `json:"items_in_progress"`
	ItemsCompleted  int     `json:"items_completed"`
	CompletionRate  float64 `json:"completion_rate"`
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
		log.Printf("Warning: Failed to get production status: %v. Using defaults.", err)
		productionStatus = &ProductionStatus{}
	}
	dashboard.ProductionStatus = productionStatus

	// Get delivery status
	deliveryStatus, err := s.getDeliveryStatus(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get delivery status: %v. Using defaults.", err)
		deliveryStatus = &DeliveryStatus{}
	}
	dashboard.DeliveryStatus = deliveryStatus

	// Get pickup status
	pickupStatus, err := s.getPickupStatus(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get pickup status: %v. Using defaults.", err)
		pickupStatus = &PickupStatus{}
	}
	dashboard.PickupStatus = pickupStatus

	// Get cleaning status
	cleaningStatus, err := s.getCleaningStatus(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get cleaning status: %v. Using defaults.", err)
		cleaningStatus = &CleaningStatus{}
	}
	dashboard.CleaningStatus = cleaningStatus

	// Get critical stock
	criticalStock, err := s.getCriticalStock(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get critical stock: %v. Using empty list.", err)
		criticalStock = []CriticalStockItem{}
	}
	dashboard.CriticalStock = criticalStock

	// Get production details (school-level)
	productionDetails, err := s.getProductionDetails(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get production details: %v. Using empty list.", err)
		productionDetails = []SchoolDetail{}
	}
	dashboard.ProductionDetails = productionDetails

	// Get delivery details (school-level)
	deliveryDetails, err := s.getDeliveryDetails(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get delivery details: %v. Using empty list.", err)
		deliveryDetails = []SchoolDetail{}
	}
	dashboard.DeliveryDetails = deliveryDetails

	// Get pickup details (school-level, stages 10-13)
	pickupDetails, err := s.getPickupDetails(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get pickup details: %v. Using empty list.", err)
		pickupDetails = []SchoolDetail{}
	}
	dashboard.PickupDetails = pickupDetails

	// Get cleaning details (school-level)
	cleaningDetails, err := s.getCleaningDetails(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get cleaning details: %v. Using empty list.", err)
		cleaningDetails = []SchoolDetail{}
	}
	dashboard.CleaningDetails = cleaningDetails

	// Calculate today's KPIs
	todayKPIs, err := s.calculateTodayKPIs(ctx)
	if err != nil {
		log.Printf("Warning: Failed to calculate KPIs: %v. Using defaults.", err)
		todayKPIs = &TodayKPIs{}
	}
	dashboard.TodayKPIs = todayKPIs

	return dashboard, nil
}

// getDummyKepalaSSPGDashboard returns dummy data for development/testing
func (s *DashboardService) getDummyKepalaSSPGDashboard() *KepalaSSPGDashboard {
	return &KepalaSSPGDashboard{
		UpdatedAt: time.Now(),
		ProductionStatus: &ProductionStatus{
			TotalRecipes:      12,
			RecipesPending:    2,
			RecipesCooking:    5,
			RecipesReady:      5,
			PackingPending:    2,
			PackingInProgress: 3,
			PackingReady:      7,
			CompletionRate:    58.3,
		},
		DeliveryStatus: &DeliveryStatus{
			TotalDeliveries: 15,
			StatusBreakdown: []StatusCount{
				{Status: "siap_dikirim", StatusLabel: "Siap Dikirim", Count: 3},
				{Status: "diperjalanan", StatusLabel: "Diperjalanan", Count: 5},
				{Status: "sudah_sampai_sekolah", StatusLabel: "Sudah Sampai Sekolah", Count: 4},
				{Status: "sudah_diterima_pihak_sekolah", StatusLabel: "Sudah Diterima", Count: 3},
			},
			CompletionRate: 20.0,
		},
		CleaningStatus: &CleaningStatus{
			TotalItems:      10,
			ItemsPending:    2,
			ItemsInProgress: 3,
			ItemsCompleted:  5,
			CompletionRate:  50.0,
		},
		CriticalStock: []CriticalStockItem{
			{
				IngredientID:   1,
				IngredientName: "Beras Putih",
				CurrentStock:   50,
				MinThreshold:   100,
				Unit:           "kg",
				DaysRemaining:  2.5,
			},
			{
				IngredientID:   2,
				IngredientName: "Minyak Goreng",
				CurrentStock:   20,
				MinThreshold:   50,
				Unit:           "liter",
				DaysRemaining:  1.8,
			},
			{
				IngredientID:   3,
				IngredientName: "Telur Ayam",
				CurrentStock:   200,
				MinThreshold:   500,
				Unit:           "butir",
				DaysRemaining:  3.2,
			},
		},
		TodayKPIs: &TodayKPIs{
			PortionsPrepared:   3250,
			DeliveryRate:       78.5,
			StockAvailability:  85.2,
			OnTimeDeliveryRate: 92.3,
		},
	}
}

// Status string groupings — source of truth for categorization
// (stage numbers in old data may be inconsistent)
var productionStatuses = []string{
	"pending", "sedang_dimasak", "selesai_dimasak",
	"siap_dipacking", "selesai_dipacking",
}

var deliveryStatuses = []string{
	"siap_dikirim", "diperjalanan",
	"sudah_sampai_sekolah", "sudah_diterima_pihak_sekolah",
}

var pickupStatuses = []string{
	"driver_menuju_lokasi_pengambilan", "driver_tiba_di_lokasi_pengambilan",
	"driver_kembali_ke_sppg", "driver_tiba_di_sppg",
}

// deliveryAndPickupStatuses combines delivery + pickup for the unified dashboard section
var deliveryAndPickupStatuses = []string{
	// Delivery statuses (stages 6-9)
	"siap_dikirim", "diperjalanan",
	"sudah_sampai_sekolah", "sudah_diterima_pihak_sekolah",
	// Pickup statuses (stages 10-13)
	"driver_menuju_lokasi_pengambilan", "driver_tiba_di_lokasi_pengambilan",
	"driver_kembali_ke_sppg", "driver_tiba_di_sppg",
}

// getProductionStatus retrieves production status for today (stages 1-5)
func (s *DashboardService) getProductionStatus(ctx context.Context) (*ProductionStatus, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	type StatusCount struct {
		Status string
		Count  int64
	}
	
	var statusCounts []StatusCount
	err := s.db.WithContext(ctx).
		Table("delivery_records").
		Select("current_status as status, COUNT(*) as count").
		Where("delivery_date >= ? AND delivery_date < ?", today, tomorrow).
		Where("current_status IN ?", productionStatuses).
		Group("current_status").
		Scan(&statusCounts).Error
	
	if err != nil {
		log.Printf("Error querying production status: %v", err)
		return nil, err
	}
	
	var totalRecipes, pending, cooking, ready, packingPending, packingInProgress, packingReady int
	
	for _, sc := range statusCounts {
		totalRecipes += int(sc.Count)
		
		switch sc.Status {
		case "pending":
			pending += int(sc.Count)
		case "sedang_dimasak":
			cooking += int(sc.Count)
		case "selesai_dimasak":
			ready += int(sc.Count)
		case "siap_dipacking", "siap_packing":
			packingPending += int(sc.Count)
		case "sedang_packing":
			packingInProgress += int(sc.Count)
		case "selesai_dipacking":
			packingReady += int(sc.Count)
		default:
			pending += int(sc.Count)
		}
	}

	log.Printf("Dashboard: Found %d production records (pending: %d, cooking: %d, ready: %d, packing: %d/%d/%d)", 
		totalRecipes, pending, cooking, ready, packingPending, packingInProgress, packingReady)

	completionRate := 0.0
	if totalRecipes > 0 {
		completionRate = roundToDecimal((float64(packingReady)/float64(totalRecipes))*100, 2)
	}

	return &ProductionStatus{
		TotalRecipes:      totalRecipes,
		RecipesPending:    pending,
		RecipesCooking:    cooking,
		RecipesReady:      ready,
		PackingPending:    packingPending,
		PackingInProgress: packingInProgress,
		PackingReady:      packingReady,
		CompletionRate:    completionRate,
	}, nil
}

// getDeliveryStatus retrieves delivery & pickup status for today (stages 6-13)
func (s *DashboardService) getDeliveryStatus(ctx context.Context) (*DeliveryStatus, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	type DBStatusCount struct {
		Status string
		Count  int64
	}
	
	var dbCounts []DBStatusCount
	err := s.db.WithContext(ctx).
		Table("delivery_records").
		Select("current_status as status, COUNT(*) as count").
		Where("delivery_date >= ? AND delivery_date < ?", today, tomorrow).
		Where("current_status IN ?", deliveryAndPickupStatuses).
		Group("current_status").
		Scan(&dbCounts).Error
	
	if err != nil {
		log.Printf("Error querying delivery status: %v", err)
		return nil, err
	}

	// Build status breakdown with labels
	var statusBreakdown []StatusCount
	var total, completed int
	
	for _, sc := range dbCounts {
		label := mapDeliveryAndPickupStatus(sc.Status)
		statusBreakdown = append(statusBreakdown, StatusCount{
			Status:      sc.Status,
			StatusLabel: label,
			Count:       int(sc.Count),
		})
		total += int(sc.Count)
		// Completed = sudah_diterima or driver_tiba_di_sppg
		if sc.Status == "sudah_diterima_pihak_sekolah" || sc.Status == "driver_tiba_di_sppg" {
			completed += int(sc.Count)
		}
	}

	completionRate := 0.0
	if total > 0 {
		completionRate = roundToDecimal((float64(completed)/float64(total))*100, 2)
	}

	log.Printf("Dashboard: Delivery & Pickup status - total: %d, breakdown: %v", total, statusBreakdown)

	return &DeliveryStatus{
		TotalDeliveries: total,
		StatusBreakdown: statusBreakdown,
		CompletionRate:  completionRate,
	}, nil
}

// getPickupStatus retrieves ompreng pickup status for today (stages 10-13)
func (s *DashboardService) getPickupStatus(ctx context.Context) (*PickupStatus, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	type DBStatusCount struct {
		Status string
		Count  int64
	}

	var dbCounts []DBStatusCount
	err := s.db.WithContext(ctx).
		Table("delivery_records").
		Select("current_status as status, COUNT(*) as count").
		Where("delivery_date >= ? AND delivery_date < ?", today, tomorrow).
		Where("current_status IN ?", pickupStatuses).
		Group("current_status").
		Scan(&dbCounts).Error

	if err != nil {
		log.Printf("Error querying pickup status: %v", err)
		return nil, err
	}

	var statusBreakdown []StatusCount
	var total, completed int
	
	for _, sc := range dbCounts {
		label := mapPickupStatus(sc.Status)
		statusBreakdown = append(statusBreakdown, StatusCount{
			Status:      sc.Status,
			StatusLabel: label,
			Count:       int(sc.Count),
		})
		total += int(sc.Count)
		if sc.Status == "driver_tiba_di_sppg" {
			completed += int(sc.Count)
		}
	}

	completionRate := 0.0
	if total > 0 {
		completionRate = roundToDecimal((float64(completed)/float64(total))*100, 2)
	}

	log.Printf("Dashboard: Pickup status - total: %d, breakdown: %v", total, statusBreakdown)

	return &PickupStatus{
		TotalPickups:    total,
		StatusBreakdown: statusBreakdown,
		CompletionRate:  completionRate,
	}, nil
}

// getCleaningStatus retrieves cleaning status for today
func (s *DashboardService) getCleaningStatus(ctx context.Context) (*CleaningStatus, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	todayStr := today.Format("2006-01-02")
	
	log.Printf("Dashboard: Getting cleaning status for date: %s", todayStr)
	
	var pending, inProgress, completed int
	
	// Try to get cleaning activities from database first
	type CleaningCount struct {
		Status string
		Count  int64
	}
	
	var counts []CleaningCount
	err := s.db.WithContext(ctx).
		Table("ompreng_cleanings").
		Select("cleaning_status as status, COUNT(*) as count").
		Joins("JOIN delivery_records ON ompreng_cleanings.delivery_record_id = delivery_records.id").
		Where("delivery_records.delivery_date >= ? AND delivery_records.delivery_date < ?", today, tomorrow).
		Group("cleaning_status").
		Scan(&counts).Error
	
	if err != nil {
		log.Printf("Dashboard: Database query failed: %v", err)
		// Don't return error, try Firebase as fallback
	} else {
		log.Printf("Dashboard: Found %d cleaning status groups from database", len(counts))
		for _, count := range counts {
			log.Printf("Dashboard: Cleaning status %s: %d items", count.Status, count.Count)
			switch count.Status {
			case "pending":
				pending = int(count.Count)
			case "in_progress":
				inProgress = int(count.Count)
			case "completed":
				completed = int(count.Count)
			}
		}
	}
	
	// Also check Firebase for any additional data (in case database is not synced)
	if s.dbClient != nil {
		log.Printf("Dashboard: Checking Firebase for cleaning data...")
		cleaningPath := "/cleaning/pending"
		var cleaningRecords map[string]interface{}
		err := s.dbClient.NewRef(cleaningPath).Get(ctx, &cleaningRecords)
		
		if err != nil {
			if err.Error() != "client: no data at ref" {
				log.Printf("Warning: Failed to get cleaning status from Firebase: %v", err)
			}
		} else if cleaningRecords != nil {
			log.Printf("Dashboard: Found %d cleaning records in Firebase", len(cleaningRecords))
			
			// Count Firebase records (only if database had no data)
			if pending == 0 && inProgress == 0 && completed == 0 {
				for key, v := range cleaningRecords {
					if cleaningData, ok := v.(map[string]interface{}); ok {
						deliveryDate, _ := cleaningData["delivery_date"].(string)
						status, _ := cleaningData["status"].(string)
						
						log.Printf("Dashboard: Cleaning record %s - delivery_date: %s, status: %s", key, deliveryDate, status)
						
						if deliveryDate == todayStr {
							switch status {
							case "pending":
								pending++
							case "in_progress":
								inProgress++
							case "completed":
								completed++
							}
						}
					}
				}
			}
		}
	}

	total := pending + inProgress + completed
	completionRate := 0.0
	if total > 0 {
		completionRate = roundToDecimal((float64(completed)/float64(total))*100, 2)
	}

	log.Printf("Dashboard: Cleaning status for %s - total: %d (pending: %d, in_progress: %d, completed: %d, rate: %.2f%%)", 
		todayStr, total, pending, inProgress, completed, completionRate)

	return &CleaningStatus{
		TotalItems:      total,
		ItemsPending:    pending,
		ItemsInProgress: inProgress,
		ItemsCompleted:  completed,
		CompletionRate:  completionRate,
	}, nil
}

// getCriticalStock retrieves items below minimum threshold
func (s *DashboardService) getCriticalStock(ctx context.Context) ([]CriticalStockItem, error) {
	alerts, err := s.inventoryService.CheckLowStock()
	if err != nil {
		log.Printf("Error checking low stock: %v", err)
		return nil, err
	}

	log.Printf("Dashboard: Found %d critical stock items", len(alerts))

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

// getProductionDetails retrieves school-level production details for today (stages 1-5)
func (s *DashboardService) getProductionDetails(ctx context.Context) ([]SchoolDetail, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	type SchoolProduction struct {
		SchoolID   uint
		SchoolName string
		Portions   int
		Status     string
	}
	
	var schoolProductions []SchoolProduction
	err := s.db.WithContext(ctx).
		Table("delivery_records").
		Select("schools.id as school_id, schools.name as school_name, delivery_records.portions as portions, delivery_records.current_status as status").
		Joins("JOIN schools ON delivery_records.school_id = schools.id").
		Where("delivery_records.delivery_date >= ? AND delivery_records.delivery_date < ?", today, tomorrow).
		Where("delivery_records.current_status IN ?", productionStatuses).
		Scan(&schoolProductions).Error
	
	if err != nil {
		log.Printf("Error querying production details: %v", err)
		return nil, err
	}
	
	details := make([]SchoolDetail, 0, len(schoolProductions))
	for _, sp := range schoolProductions {
		details = append(details, SchoolDetail{
			SchoolID:   sp.SchoolID,
			SchoolName: sp.SchoolName,
			Portions:   sp.Portions,
			Status:     mapProductionStatus(sp.Status, 0),
		})
	}
	
	log.Printf("Dashboard: Found %d production details", len(details))
	return details, nil
}

// getDeliveryDetails retrieves school-level delivery & pickup details for today (stages 6-13)
func (s *DashboardService) getDeliveryDetails(ctx context.Context) ([]SchoolDetail, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	type SchoolDelivery struct {
		SchoolID   uint
		SchoolName string
		Portions   int
		Status     string
	}
	
	var schoolDeliveries []SchoolDelivery
	err := s.db.WithContext(ctx).
		Table("delivery_records").
		Select("schools.id as school_id, schools.name as school_name, delivery_records.portions as portions, delivery_records.current_status as status").
		Joins("JOIN schools ON delivery_records.school_id = schools.id").
		Where("delivery_records.delivery_date >= ? AND delivery_records.delivery_date < ?", today, tomorrow).
		Where("delivery_records.current_status IN ?", deliveryAndPickupStatuses).
		Scan(&schoolDeliveries).Error
	
	if err != nil {
		log.Printf("Error querying delivery details: %v", err)
		return nil, err
	}
	
	details := make([]SchoolDetail, 0, len(schoolDeliveries))
	for _, sd := range schoolDeliveries {
		details = append(details, SchoolDetail{
			SchoolID:   sd.SchoolID,
			SchoolName: sd.SchoolName,
			Portions:   sd.Portions,
			Status:     mapDeliveryAndPickupStatus(sd.Status),
		})
	}
	
	log.Printf("Dashboard: Found %d delivery & pickup details", len(details))
	return details, nil
}

// getPickupDetails retrieves school-level ompreng pickup details for today (stages 10-13)
func (s *DashboardService) getPickupDetails(ctx context.Context) ([]SchoolDetail, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	type SchoolPickup struct {
		SchoolID   uint
		SchoolName string
		OmprengCount int
		Status     string
	}

	var schoolPickups []SchoolPickup
	err := s.db.WithContext(ctx).
		Table("delivery_records").
		Select("schools.id as school_id, schools.name as school_name, delivery_records.ompreng_count, delivery_records.current_status as status").
		Joins("JOIN schools ON delivery_records.school_id = schools.id").
		Where("delivery_records.delivery_date >= ? AND delivery_records.delivery_date < ?", today, tomorrow).
		Where("delivery_records.current_status IN ?", pickupStatuses).
		Scan(&schoolPickups).Error

	if err != nil {
		log.Printf("Error querying pickup details: %v", err)
		return nil, err
	}

	details := make([]SchoolDetail, 0, len(schoolPickups))
	for _, sp := range schoolPickups {
		details = append(details, SchoolDetail{
			SchoolID:   sp.SchoolID,
			SchoolName: sp.SchoolName,
			Portions:   sp.OmprengCount,
			Status:     mapPickupStatus(sp.Status),
		})
	}

	log.Printf("Dashboard: Found %d pickup details", len(details))
	return details, nil
}

// getCleaningDetails retrieves school-level cleaning details for today
func (s *DashboardService) getCleaningDetails(ctx context.Context) ([]SchoolDetail, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	type SchoolCleaning struct {
		SchoolID       uint
		SchoolName     string
		OmprengCount   int
		CleaningStatus string
	}
	
	var schoolCleanings []SchoolCleaning
	err := s.db.WithContext(ctx).
		Table("ompreng_cleanings").
		Select("schools.id as school_id, schools.name as school_name, ompreng_cleanings.ompreng_count, ompreng_cleanings.cleaning_status").
		Joins("JOIN delivery_records ON ompreng_cleanings.delivery_record_id = delivery_records.id").
		Joins("JOIN schools ON delivery_records.school_id = schools.id").
		Where("delivery_records.delivery_date >= ? AND delivery_records.delivery_date < ?", today, tomorrow).
		Scan(&schoolCleanings).Error
	
	if err != nil {
		log.Printf("Error querying cleaning details: %v", err)
		return nil, err
	}
	
	// Convert to SchoolDetail format
	details := make([]SchoolDetail, 0, len(schoolCleanings))
	for _, sc := range schoolCleanings {
		// Map status to user-friendly text
		statusText := mapCleaningStatus(sc.CleaningStatus)
		
		details = append(details, SchoolDetail{
			SchoolID:   sc.SchoolID,
			SchoolName: sc.SchoolName,
			Portions:   sc.OmprengCount,
			Status:     statusText,
		})
	}
	
	log.Printf("Dashboard: Found %d cleaning details", len(details))
	return details, nil
}

// mapProductionStatus maps production status to user-friendly text
func mapProductionStatus(status string, stage int) string {
	switch status {
	case "pending":
		return "Menunggu"
	case "sedang_dimasak":
		return "Sedang Dimasak"
	case "selesai_dimasak":
		return "Selesai Dimasak"
	case "siap_dipacking":
		return "Siap Packing"
	case "siap_packing":
		return "Siap Packing"
	case "sedang_packing":
		return "Sedang Packing"
	case "selesai_dipacking":
		return "Selesai Packing"
	case "siap_dikirim":
		return "Siap Dikirim"
	default:
		// Fallback based on stage
		if stage <= 1 {
			return "Menunggu"
		} else if stage <= 2 {
			return "Sedang Dimasak"
		} else if stage <= 3 {
			return "Selesai Dimasak"
		} else if stage <= 4 {
			return "Sedang Packing"
		}
		return "Selesai"
	}
}

// mapDeliveryStatus maps delivery status to user-friendly text (stages 6-9)
func mapDeliveryStatus(status string, stage int) string {
	switch status {
	case "siap_dikirim":
		return "Siap Dikirim"
	case "diperjalanan", "dalam_perjalanan":
		return "Diperjalanan"
	case "sudah_sampai_sekolah", "tiba_di_sekolah":
		return "Sudah Sampai Sekolah"
	case "sudah_diterima_pihak_sekolah":
		return "Sudah Diterima"
	case "selesai":
		return "Selesai"
	default:
		return "Dalam Perjalanan"
	}
}

// mapDeliveryAndPickupStatus maps both delivery and pickup statuses to user-friendly text
func mapDeliveryAndPickupStatus(status string) string {
	switch status {
	// Delivery statuses
	case "siap_dikirim":
		return "Siap Dikirim"
	case "diperjalanan", "dalam_perjalanan":
		return "Diperjalanan"
	case "sudah_sampai_sekolah", "tiba_di_sekolah":
		return "Sudah Sampai Sekolah"
	case "sudah_diterima_pihak_sekolah":
		return "Sudah Diterima"
	// Pickup statuses
	case "driver_menuju_lokasi_pengambilan":
		return "Driver Menuju Lokasi Pengambilan"
	case "driver_tiba_di_lokasi_pengambilan":
		return "Driver Tiba di Lokasi"
	case "driver_kembali_ke_sppg":
		return "Driver Kembali ke SPPG"
	case "driver_tiba_di_sppg":
		return "Driver Tiba di SPPG"
	default:
		return status
	}
}

// mapPickupStatus maps pickup status to user-friendly text (stages 10-13)
func mapPickupStatus(status string) string {
	switch status {
	case "driver_menuju_lokasi_pengambilan":
		return "Driver Menuju Lokasi Pengambilan"
	case "driver_tiba_di_lokasi_pengambilan":
		return "Driver Tiba di Lokasi"
	case "driver_kembali_ke_sppg":
		return "Driver Kembali ke SPPG"
	case "driver_tiba_di_sppg":
		return "Driver Tiba di SPPG"
	default:
		return "Dalam Proses Pengambilan"
	}
}

// mapCleaningStatus maps cleaning status to user-friendly text
func mapCleaningStatus(status string) string {
	switch status {
	case "pending":
		return "Menunggu"
	case "in_progress":
		return "Sedang Dicuci"
	case "completed":
		return "Selesai"
	default:
		return "Menunggu"
	}
}

// calculateTodayKPIs calculates key performance indicators for today
func (s *DashboardService) calculateTodayKPIs(ctx context.Context) (*TodayKPIs, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)
	
	// Calculate portions prepared (from menu items)
	var portionsPrepared int64
	err := s.db.WithContext(ctx).
		Table("menu_items").
		Joins("JOIN menu_plans ON menu_items.menu_plan_id = menu_plans.id").
		Where("menu_plans.status = ?", "approved").
		Where("menu_items.date >= ? AND menu_items.date < ?", today, tomorrow).
		Select("COALESCE(SUM(portions), 0)").
		Scan(&portionsPrepared).Error
	if err != nil {
		log.Printf("Error calculating portions prepared: %v", err)
		return nil, err
	}

	log.Printf("Dashboard: Portions prepared today: %d", portionsPrepared)

	// Calculate delivery rate
	tasks, err := s.deliveryTaskService.GetAllDeliveryTasks(nil, "", &today)
	if err != nil {
		log.Printf("Error getting delivery tasks for KPIs: %v", err)
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
		deliveryRate = roundToDecimal((float64(completed)/float64(len(tasks)))*100, 2)
	}

	log.Printf("Dashboard: Delivery rate: %.2f%% (%d completed out of %d)", deliveryRate, int(deliveryRate*float64(len(tasks))/100), len(tasks))

	// Calculate stock availability (percentage of items above threshold)
	allInventory, err := s.inventoryService.GetAllInventory()
	if err != nil {
		log.Printf("Error getting inventory for KPIs: %v", err)
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
		stockAvailability = roundToDecimal((float64(aboveThreshold)/float64(len(allInventory)))*100, 2)
	}

	log.Printf("Dashboard: Stock availability: %.2f%% (%d items above threshold)", stockAvailability, int(stockAvailability*float64(len(allInventory))/100))

	// Calculate on-time delivery rate from completed deliveries with POD
	onTimeDeliveryRate := 0.0
	completedCount := 0
	for _, task := range tasks {
		if task.Status == "completed" {
			completedCount++
		}
	}
	if completedCount > 0 {
		// For now, assume all completed deliveries are on-time
		// In production, this would compare actual vs expected delivery times
		onTimeDeliveryRate = 95.0
	}

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
		log.Printf("Warning: Failed to get budget absorption: %v. Using defaults.", err)
		budgetAbsorption = &BudgetAbsorption{CategoryBreakdown: []BudgetCategoryBreakdown{}}
	}
	dashboard.BudgetAbsorption = budgetAbsorption

	// Get nutrition distribution
	nutritionDistribution, err := s.getNutritionDistribution(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Warning: Failed to get nutrition distribution: %v. Using defaults.", err)
		nutritionDistribution = &NutritionDistribution{}
	}
	dashboard.NutritionDistribution = nutritionDistribution

	// Get supplier performance
	supplierPerformance, err := s.getSupplierPerformance(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get supplier performance: %v. Using defaults.", err)
		supplierPerformance = &SupplierMetrics{}
	}
	dashboard.SupplierPerformance = supplierPerformance

	// Get monthly trend
	monthlyTrend, err := s.getMonthlyTrend(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Warning: Failed to get monthly trend: %v. Using empty list.", err)
		monthlyTrend = []MonthlyMetrics{}
	}
	dashboard.MonthlyTrend = monthlyTrend

	return dashboard, nil
}

// getDummyKepalaYayasanDashboard returns dummy data for Kepala Yayasan dashboard
func (s *DashboardService) getDummyKepalaYayasanDashboard() *KepalaYayasanDashboard {
	return &KepalaYayasanDashboard{
		UpdatedAt: time.Now(),
		BudgetAbsorption: &BudgetAbsorption{
			TotalBudget:    5000000000,
			TotalSpent:     3750000000,
			AbsorptionRate: 75.0,
			CategoryBreakdown: []BudgetCategoryBreakdown{
				{Category: "bahan_baku", Budget: 3000000000, Spent: 2400000000, AbsorptionRate: 80.0},
				{Category: "gaji", Budget: 1200000000, Spent: 900000000, AbsorptionRate: 75.0},
				{Category: "operasional", Budget: 500000000, Spent: 300000000, AbsorptionRate: 60.0},
				{Category: "utilitas", Budget: 300000000, Spent: 150000000, AbsorptionRate: 50.0},
			},
		},
		NutritionDistribution: &NutritionDistribution{
			TotalPortionsDistributed: 45000,
			SchoolsServed:            15,
			StudentsReached:          3250,
			AveragePortionsPerSchool: 3000,
		},
		SupplierPerformance: &SupplierMetrics{
			TotalSuppliers:    12,
			ActiveSuppliers:   10,
			AvgOnTimeDelivery: 88.5,
			AvgQualityRating:  4.2,
		},
		MonthlyTrend: []MonthlyMetrics{
			{Month: "Januari", Year: 2026, PortionsDistributed: 42000, BudgetSpent: 350000000, SchoolsServed: 14},
			{Month: "Februari", Year: 2026, PortionsDistributed: 45000, BudgetSpent: 375000000, SchoolsServed: 15},
			{Month: "Maret", Year: 2026, PortionsDistributed: 43000, BudgetSpent: 360000000, SchoolsServed: 15},
			{Month: "April", Year: 2026, PortionsDistributed: 46000, BudgetSpent: 380000000, SchoolsServed: 16},
			{Month: "Mei", Year: 2026, PortionsDistributed: 48000, BudgetSpent: 400000000, SchoolsServed: 16},
			{Month: "Juni", Year: 2026, PortionsDistributed: 47000, BudgetSpent: 390000000, SchoolsServed: 16},
		},
	}
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
			absorptionRate = roundToDecimal((spent/budget)*100, 2)
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
		overallAbsorptionRate = roundToDecimal((totalSpent/totalBudget)*100, 2)
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

	// Calculate average portions per school (rounded to 2 decimal places)
	avgPortionsPerSchool := 0.0
	if schoolsServed > 0 {
		avgPortionsPerSchool = roundToDecimal(float64(totalPortions)/float64(schoolsServed), 2)
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
	if s.dbClient == nil {
		return fmt.Errorf("Firebase client tidak tersedia")
	}

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
	if s.dbClient == nil {
		return fmt.Errorf("Firebase client tidak tersedia")
	}

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

// ClearFirebaseKDSData clears all KDS-related data from Firebase
func (s *DashboardService) ClearFirebaseKDSData(ctx context.Context) error {
	if s.firebaseApp == nil {
		return fmt.Errorf("Firebase app tidak tersedia")
	}

	// Create Firebase sync service
	firebaseSync, err := NewFirebaseSyncService(s.firebaseApp)
	if err != nil {
		return fmt.Errorf("gagal membuat Firebase sync service: %w", err)
	}

	// Clear KDS data
	if err := firebaseSync.ClearKDSData(ctx); err != nil {
		return fmt.Errorf("gagal menghapus data KDS dari Firebase: %w", err)
	}

	return nil
}
