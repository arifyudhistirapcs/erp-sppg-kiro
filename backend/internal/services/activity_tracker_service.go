package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// ActivityTrackerService handles business logic for activity tracking operations
type ActivityTrackerService struct {
	db *gorm.DB
}

// NewActivityTrackerService creates a new ActivityTrackerService instance
func NewActivityTrackerService(db *gorm.DB) *ActivityTrackerService {
	return &ActivityTrackerService{
		db: db,
	}
}

// OrderResponse represents a simplified order for list view
type OrderResponse struct {
	ID           uint                  `json:"id"`
	OrderDate    time.Time             `json:"order_date"`
	School       SchoolInfo            `json:"school"`
	Menu         MenuInfo              `json:"menu"`
	Driver       DriverInfo            `json:"driver"`
	Portions     int                   `json:"portions"`
	CurrentStatus string               `json:"current_status"`
	CurrentStage int                   `json:"current_stage"`
	OmprengCount int                   `json:"ompreng_count"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// SchoolInfo represents school information
type SchoolInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// MenuInfo represents menu information
type MenuInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	PhotoURL string `json:"photo_url"`
}

// DriverInfo represents driver information
type DriverInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	VehicleInfo string `json:"vehicle_info"`
}

// OrderSummary represents summary statistics for orders
type OrderSummary struct {
	TotalOrders        int            `json:"total_orders"`
	StatusDistribution map[string]int `json:"status_distribution"`
}

// GetOrdersByDateResponse represents the response for GetOrdersByDate
type GetOrdersByDateResponse struct {
	Orders  []OrderResponse `json:"orders"`
	Summary OrderSummary    `json:"summary"`
}

// GetOrdersByDate retrieves all order records for a specific date with optional filters
func (s *ActivityTrackerService) GetOrdersByDate(ctx context.Context, date time.Time, schoolID *uint, search string) (*GetOrdersByDateResponse, error) {
	var deliveryRecords []models.DeliveryRecord
	
	// Normalize date to start of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	// Build query
	query := s.db.WithContext(ctx).
		Preload("School").
		Preload("Driver").
		Preload("MenuItem.Recipe").
		Where("delivery_date >= ? AND delivery_date < ?", startOfDay, endOfDay)
	
	// Apply school filter if provided
	if schoolID != nil {
		query = query.Where("school_id = ?", *schoolID)
	}
	
	// Apply search filter if provided
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("LEFT JOIN schools ON schools.id = delivery_records.school_id").
			Joins("LEFT JOIN menu_items ON menu_items.id = delivery_records.menu_item_id").
			Where("schools.name ILIKE ? OR menu_items.name ILIKE ?", searchPattern, searchPattern)
	}
	
	// Execute query
	if err := query.Order("delivery_date DESC").Find(&deliveryRecords).Error; err != nil {
		log.Printf("Error fetching orders by date: %v", err)
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	
	// Build response
	orders := make([]OrderResponse, len(deliveryRecords))
	statusDistribution := make(map[string]int)
	
	for i, record := range deliveryRecords {
		orders[i] = OrderResponse{
			ID:        record.ID,
			OrderDate: record.DeliveryDate,
			School: SchoolInfo{
				ID:      record.School.ID,
				Name:    record.School.Name,
				Address: record.School.Address,
			},
			Menu: MenuInfo{
				ID:       record.MenuItem.ID,
				Name:     record.MenuItem.Recipe.Name,
				PhotoURL: record.MenuItem.Recipe.PhotoURL,
			},
			Driver: DriverInfo{
				ID:          record.Driver.ID,
				Name:        record.Driver.FullName,
				VehicleInfo: "", // TODO: Add vehicle info to User model or create separate Driver model
			},
			Portions:      record.Portions,
			CurrentStatus: record.CurrentStatus,
			CurrentStage:  record.CurrentStage,
			OmprengCount:  record.OmprengCount,
			UpdatedAt:     record.UpdatedAt,
		}
		
		// Count status distribution
		statusDistribution[record.CurrentStatus]++
	}
	
	return &GetOrdersByDateResponse{
		Orders: orders,
		Summary: OrderSummary{
			TotalOrders:        len(orders),
			StatusDistribution: statusDistribution,
		},
	}, nil
}

// TimelineStage represents a single stage in the order timeline
type TimelineStage struct {
	Stage          int        `json:"stage"`
	Status         string     `json:"status"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	IsCompleted    bool       `json:"is_completed"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	TransitionedBy *UserInfo  `json:"transitioned_by"`
	Media          *MediaInfo `json:"media"`
}

// UserInfo represents user information
type UserInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// MediaInfo represents media information
type MediaInfo struct {
	Type         string `json:"type"` // "photo" or "video"
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// OrderDetailResponse represents detailed order information with timeline
type OrderDetailResponse struct {
	ID           uint            `json:"id"`
	OrderDate    time.Time       `json:"order_date"`
	School       SchoolDetailInfo `json:"school"`
	Driver       DriverDetailInfo `json:"driver"`
	Menu         MenuInfo        `json:"menu"`
	Portions     int             `json:"portions"`
	CurrentStatus string         `json:"current_status"`
	CurrentStage int             `json:"current_stage"`
	OmprengCount int             `json:"ompreng_count"`
	Timeline     []TimelineStage `json:"timeline"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// SchoolDetailInfo represents detailed school information
type SchoolDetailInfo struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	ContactPerson string `json:"contact_person"`
	PhoneNumber   string `json:"phone_number"`
}

// DriverDetailInfo represents detailed driver information
type DriverDetailInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	VehicleInfo string `json:"vehicle_info"`
}

// GetOrderDetails retrieves detailed information including vertical timeline data
func (s *ActivityTrackerService) GetOrderDetails(ctx context.Context, orderID uint) (*OrderDetailResponse, error) {
	var deliveryRecord models.DeliveryRecord
	
	// Fetch delivery record with relations
	if err := s.db.WithContext(ctx).
		Preload("School").
		Preload("Driver").
		Preload("MenuItem.Recipe").
		First(&deliveryRecord, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		log.Printf("Error fetching order details: %v", err)
		return nil, fmt.Errorf("failed to fetch order details: %w", err)
	}
	
	// Fetch status transitions
	var transitions []models.StatusTransition
	if err := s.db.WithContext(ctx).
		Preload("User").
		Where("delivery_record_id = ?", orderID).
		Order("stage ASC, transitioned_at ASC").
		Find(&transitions).Error; err != nil {
		log.Printf("Error fetching status transitions: %v", err)
		return nil, fmt.Errorf("failed to fetch status transitions: %w", err)
	}
	
	// Build timeline with all 16 stages
	timeline := s.buildTimeline(deliveryRecord.CurrentStage, transitions)
	
	// Build response
	response := &OrderDetailResponse{
		ID:        deliveryRecord.ID,
		OrderDate: deliveryRecord.DeliveryDate,
		School: SchoolDetailInfo{
			ID:            deliveryRecord.School.ID,
			Name:          deliveryRecord.School.Name,
			Address:       deliveryRecord.School.Address,
			ContactPerson: deliveryRecord.School.ContactPerson,
			PhoneNumber:   deliveryRecord.School.PhoneNumber,
		},
		Driver: DriverDetailInfo{
			ID:          deliveryRecord.Driver.ID,
			Name:        deliveryRecord.Driver.FullName,
			PhoneNumber: deliveryRecord.Driver.PhoneNumber,
			VehicleInfo: "", // TODO: Add vehicle info
		},
		Menu: MenuInfo{
			ID:       deliveryRecord.MenuItem.ID,
			Name:     deliveryRecord.MenuItem.Recipe.Name,
			PhotoURL: deliveryRecord.MenuItem.Recipe.PhotoURL,
		},
		Portions:      deliveryRecord.Portions,
		CurrentStatus: deliveryRecord.CurrentStatus,
		CurrentStage:  deliveryRecord.CurrentStage,
		OmprengCount:  deliveryRecord.OmprengCount,
		Timeline:      timeline,
		CreatedAt:     deliveryRecord.CreatedAt,
		UpdatedAt:     deliveryRecord.UpdatedAt,
	}
	
	return response, nil
}

// buildTimeline constructs the complete 16-stage timeline
func (s *ActivityTrackerService) buildTimeline(currentStage int, transitions []models.StatusTransition) []TimelineStage {
	// Define all 16 stages
	stageDefinitions := []struct {
		Stage       int
		Status      string
		Title       string
		Description string
	}{
		{1, "order_disiapkan", "Order sedang disiapkan", "Makanan sedang diproses dan disiapkan untuk dimasak."},
		{2, "order_dimasak", "Order sedang dimasak", "Makanan sedang dimasak sesuai menu yang dijadwalkan."},
		{3, "order_dikemas", "Order sedang dikemas", "Makanan sedang dalam proses dikemas menjadi porsi satuan."},
		{4, "order_siap_diambil", "Order siap diambil", "Pesanan sudah siap di dapur dan menunggu diambil untuk dikirim ke sekolah."},
		{5, "pesanan_dalam_perjalanan", "Pesanan dalam perjalanan", "Driver sedang dalam perjalanan menuju sekolah."},
		{6, "pesanan_sudah_tiba", "Pesanan sudah tiba", "Driver sudah tiba di sekolah."},
		{7, "pesanan_sudah_diterima", "Pesanan sudah diterima", "Pesanan sudah diterima pihak sekolah."},
		{8, "driver_menuju_lokasi", "Driver menuju lokasi pengambilan", "Driver diperjalanan mengambil menu/ompreng/wadah makan."},
		{9, "driver_tiba_di_lokasi", "Driver sudah tiba di lokasi", "Driver sudah tiba di sekolah untuk mengambil menu/ompreng/wadah makan."},
		{10, "driver_kembali", "Driver dalam perjalanan kembali", "Driver diperjalanan kembali ke SPPG untuk mengembalikan menu/ompreng/wadah makan."},
		{11, "driver_tiba_di_sppg", "Driver sudah tiba di SPPG", "Driver tiba di SPPG untuk mengembalikan menu/ompreng/wadah makan."},
		{12, "ompreng_siap_dicuci", "Ompreng siap dicuci", "Menu/ompreng/wadah makan siap di cuci."},
		{13, "ompreng_sedang_dicuci", "Ompreng sedang dicuci", "Menu/ompreng/wadah makan proses di cuci."},
		{14, "ompreng_selesai_dicuci", "Ompreng selesai dicuci", "Menu/ompreng/wadah makan selesai dicuci."},
		{15, "ompreng_siap_digunakan", "Ompreng siap digunakan kembali", "Ompreng sudah bersih dan siap digunakan untuk order berikutnya."},
		{16, "order_selesai", "Order selesai", "Seluruh proses order telah selesai."},
	}
	
	// Create map of transitions by stage
	transitionMap := make(map[int]*models.StatusTransition)
	for i := range transitions {
		transitionMap[transitions[i].Stage] = &transitions[i]
	}
	
	// Build timeline
	timeline := make([]TimelineStage, 16)
	for i, def := range stageDefinitions {
		stage := TimelineStage{
			Stage:       def.Stage,
			Status:      def.Status,
			Title:       def.Title,
			Description: def.Description,
			IsCompleted: def.Stage < currentStage,
		}
		
		// Add transition data if exists
		if transition, exists := transitionMap[def.Stage]; exists {
			stage.StartedAt = &transition.TransitionedAt
			stage.CompletedAt = &transition.TransitionedAt
			stage.TransitionedBy = &UserInfo{
				ID:   transition.User.ID,
				Name: transition.User.FullName,
			}
			
			// Add media if exists
			if transition.MediaURL != "" {
				stage.Media = &MediaInfo{
					Type:         transition.MediaType,
					URL:          transition.MediaURL,
					ThumbnailURL: transition.MediaURL, // TODO: Generate actual thumbnail
				}
			}
		}
		
		timeline[i] = stage
	}
	
	return timeline
}

// UpdateOrderStatus updates order status and creates status transition
func (s *ActivityTrackerService) UpdateOrderStatus(ctx context.Context, orderID uint, newStatus string, stage int, userID uint, notes string) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Fetch current delivery record
	var deliveryRecord models.DeliveryRecord
	if err := tx.First(&deliveryRecord, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to fetch order: %w", err)
	}
	
	// Validate stage progression
	if stage < deliveryRecord.CurrentStage {
		log.Printf("Warning: Backward transition attempted from stage %d to %d for order %d", deliveryRecord.CurrentStage, stage, orderID)
		// Allow backward transitions but log warning
	}
	
	// Check for skipped stages
	if stage > deliveryRecord.CurrentStage+1 {
		log.Printf("Warning: Skipped stages detected from stage %d to %d for order %d", deliveryRecord.CurrentStage, stage, orderID)
	}
	
	// Create status transition record
	transition := models.StatusTransition{
		DeliveryRecordID: orderID,
		FromStatus:       deliveryRecord.CurrentStatus,
		ToStatus:         newStatus,
		Stage:            stage,
		TransitionedAt:   time.Now(),
		TransitionedBy:   userID,
		Notes:            notes,
	}
	
	if err := tx.Create(&transition).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating status transition: %v", err)
		return fmt.Errorf("failed to create status transition: %w", err)
	}
	
	// Update delivery record
	if err := tx.Model(&deliveryRecord).Updates(map[string]interface{}{
		"current_status": newStatus,
		"current_stage":  stage,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating delivery record: %v", err)
		return fmt.Errorf("failed to update delivery record: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Printf("Order %d status updated from %s (stage %d) to %s (stage %d)", orderID, deliveryRecord.CurrentStatus, deliveryRecord.CurrentStage, newStatus, stage)
	
	return nil
}

// AttachStageMedia attaches photo or video URL to a specific stage transition
func (s *ActivityTrackerService) AttachStageMedia(ctx context.Context, orderID uint, stage int, mediaURL string, mediaType string) error {
	// Find the status transition for this stage
	var transition models.StatusTransition
	if err := s.db.WithContext(ctx).
		Where("delivery_record_id = ? AND stage = ?", orderID, stage).
		First(&transition).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("status transition not found for stage %d", stage)
		}
		return fmt.Errorf("failed to fetch status transition: %w", err)
	}
	
	// Update media fields
	if err := s.db.WithContext(ctx).Model(&transition).Updates(map[string]interface{}{
		"media_url":  mediaURL,
		"media_type": mediaType,
	}).Error; err != nil {
		log.Printf("Error attaching media to stage %d: %v", stage, err)
		return fmt.Errorf("failed to attach media: %w", err)
	}
	
	log.Printf("Media attached to order %d stage %d: %s (%s)", orderID, stage, mediaURL, mediaType)
	
	return nil
}

// SyncOrderToFirebase synchronizes order record to Firebase
func (s *ActivityTrackerService) SyncOrderToFirebase(ctx context.Context, orderID uint, firebaseApp interface{}) error {
	// Fetch order details
	orderDetails, err := s.GetOrderDetails(ctx, orderID)
	if err != nil {
		log.Printf("Error fetching order details for Firebase sync: %v", err)
		return fmt.Errorf("failed to fetch order details: %w", err)
	}
	
	// Format date for Firebase path
	dateStr := orderDetails.OrderDate.Format("2006-01-02")
	
	// Prepare Firebase data
	firebaseData := map[string]interface{}{
		"id":             orderDetails.ID,
		"order_date":     orderDetails.OrderDate.Format(time.RFC3339),
		"school_id":      orderDetails.School.ID,
		"school_name":    orderDetails.School.Name,
		"driver_id":      orderDetails.Driver.ID,
		"driver_name":    orderDetails.Driver.Name,
		"menu_name":      orderDetails.Menu.Name,
		"menu_photo_url": orderDetails.Menu.PhotoURL,
		"portions":       orderDetails.Portions,
		"current_status": orderDetails.CurrentStatus,
		"current_stage":  orderDetails.CurrentStage,
		"ompreng_count":  orderDetails.OmprengCount,
		"updated_at":     orderDetails.UpdatedAt.Format(time.RFC3339),
	}
	
	// Prepare transitions data
	transitions := make([]map[string]interface{}, 0)
	for _, stage := range orderDetails.Timeline {
		if stage.IsCompleted && stage.TransitionedBy != nil {
			transition := map[string]interface{}{
				"stage":       stage.Stage,
				"from_status": "", // TODO: Get from previous stage
				"to_status":   stage.Status,
				"transitioned_at": stage.CompletedAt.Format(time.RFC3339),
				"transitioned_by_name": stage.TransitionedBy.Name,
				"notes": "",
			}
			
			if stage.Media != nil {
				transition["media_url"] = stage.Media.URL
				transition["media_type"] = stage.Media.Type
			}
			
			transitions = append(transitions, transition)
		}
	}
	
	// Note: Actual Firebase write would be done here
	// For now, we'll log the data that would be synced
	log.Printf("Would sync to Firebase: order_tracking/%s/%d", dateStr, orderID)
	log.Printf("Order data: %+v", firebaseData)
	log.Printf("Transitions count: %d", len(transitions))
	
	// TODO: Implement actual Firebase write
	// This requires Firebase Realtime Database client
	// Example:
	// client, err := firebaseApp.Database(ctx)
	// if err != nil {
	//     return fmt.Errorf("failed to get Firebase client: %w", err)
	// }
	// ref := client.NewRef(fmt.Sprintf("order_tracking/%s/%d", dateStr, orderID))
	// if err := ref.Set(ctx, firebaseData); err != nil {
	//     return fmt.Errorf("failed to write to Firebase: %w", err)
	// }
	
	return nil
}

// HandleKDSStatusUpdate processes status updates from KDS modules
func (s *ActivityTrackerService) HandleKDSStatusUpdate(ctx context.Context, orderID uint, kdsStatus string, userID uint) error {
	// Map KDS status to Activity Tracker status and stage
	statusMap := map[string]struct {
		status string
		stage  int
	}{
		"cooking":           {"order_dimasak", 2},
		"ready":             {"order_dikemas", 3},
		"packing_completed": {"order_siap_diambil", 4},
	}
	
	mapping, exists := statusMap[kdsStatus]
	if !exists {
		log.Printf("Warning: Unknown KDS status: %s", kdsStatus)
		return fmt.Errorf("unknown KDS status: %s", kdsStatus)
	}
	
	// Update order status
	notes := fmt.Sprintf("Status updated from KDS module: %s", kdsStatus)
	return s.UpdateOrderStatus(ctx, orderID, mapping.status, mapping.stage, userID, notes)
}

// HandleLogisticsStatusUpdate processes status updates from Logistics/Delivery module
func (s *ActivityTrackerService) HandleLogisticsStatusUpdate(ctx context.Context, orderID uint, logisticsStatus string, userID uint) error {
	// Map Logistics status to Activity Tracker status and stage
	statusMap := map[string]struct {
		status string
		stage  int
	}{
		"driver_departed":       {"pesanan_dalam_perjalanan", 5},
		"driver_arrived":        {"pesanan_sudah_tiba", 6},
		"delivery_confirmed":    {"pesanan_sudah_diterima", 7},
		"collection_assigned":   {"driver_menuju_lokasi", 8},
		"collection_arrived":    {"driver_tiba_di_lokasi", 9},
		"collection_departed":   {"driver_kembali", 10},
		"collection_completed":  {"driver_tiba_di_sppg", 11},
		"ready_for_cleaning":    {"ompreng_siap_dicuci", 12},
		"cleaning":              {"ompreng_sedang_dicuci", 13},
		"cleaning_completed":    {"ompreng_selesai_dicuci", 14},
		"ready_for_reuse":       {"ompreng_siap_digunakan", 15},
		"order_completed":       {"order_selesai", 16},
	}
	
	mapping, exists := statusMap[logisticsStatus]
	if !exists {
		log.Printf("Warning: Unknown logistics status: %s", logisticsStatus)
		return fmt.Errorf("unknown logistics status: %s", logisticsStatus)
	}
	
	// Update order status
	notes := fmt.Sprintf("Status updated from Logistics module: %s", logisticsStatus)
	return s.UpdateOrderStatus(ctx, orderID, mapping.status, mapping.stage, userID, notes)
}
