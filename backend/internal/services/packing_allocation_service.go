package services

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// PackingAllocationService handles packing allocation operations
type PackingAllocationService struct {
	db          *gorm.DB
	firebaseApp *firebase.App
	dbClient    *db.Client
}

// NewPackingAllocationService creates a new packing allocation service instance
func NewPackingAllocationService(database *gorm.DB, firebaseApp *firebase.App) (*PackingAllocationService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase database client: %w", err)
	}

	return &PackingAllocationService{
		db:          database,
		firebaseApp: firebaseApp,
		dbClient:    dbClient,
	}, nil
}

// SchoolAllocation represents packing allocation for a school
type SchoolAllocation struct {
	SchoolID   uint              `json:"school_id"`
	SchoolName string            `json:"school_name"`
	Portions   int               `json:"portions"`
	MenuItems  []MenuItemSummary `json:"menu_items"`
	Status     string            `json:"status"` // pending, packing, ready
}

// MenuItemSummary represents a menu item summary for packing
type MenuItemSummary struct {
	RecipeID   uint   `json:"recipe_id"`
	RecipeName string `json:"recipe_name"`
	Portions   int    `json:"portions"`
}

// CalculatePackingAllocations calculates portion distribution per school for today
func (s *PackingAllocationService) CalculatePackingAllocations(ctx context.Context) ([]SchoolAllocation, error) {
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get today's delivery tasks
	var deliveryTasks []models.DeliveryTask
	err := s.db.WithContext(ctx).
		Preload("School").
		Preload("MenuItems").
		Preload("MenuItems.Recipe").
		Where("task_date >= ? AND task_date < ?", startOfDay, endOfDay).
		Where("status != ?", "cancelled").
		Find(&deliveryTasks).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery tasks: %w", err)
	}

	// Group by school
	schoolMap := make(map[uint]*SchoolAllocation)
	for _, task := range deliveryTasks {
		allocation, exists := schoolMap[task.SchoolID]
		if !exists {
			allocation = &SchoolAllocation{
				SchoolID:   task.School.ID,
				SchoolName: task.School.Name,
				Portions:   0,
				MenuItems:  []MenuItemSummary{},
				Status:     "pending",
			}
			schoolMap[task.SchoolID] = allocation
		}

		// Add menu items
		for _, menuItem := range task.MenuItems {
			allocation.Portions += menuItem.Portions
			allocation.MenuItems = append(allocation.MenuItems, MenuItemSummary{
				RecipeID:   menuItem.Recipe.ID,
				RecipeName: menuItem.Recipe.Name,
				Portions:   menuItem.Portions,
			})
		}
	}

	// Convert map to slice
	allocations := make([]SchoolAllocation, 0, len(schoolMap))
	for _, allocation := range schoolMap {
		allocations = append(allocations, *allocation)
	}

	return allocations, nil
}

// GetPackingAllocations retrieves packing allocations for today
func (s *PackingAllocationService) GetPackingAllocations(ctx context.Context) ([]SchoolAllocation, error) {
	return s.CalculatePackingAllocations(ctx)
}

// UpdatePackingStatus updates the packing status for a school
func (s *PackingAllocationService) UpdatePackingStatus(ctx context.Context, schoolID uint, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"pending": true,
		"packing": true,
		"ready":   true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get school details
	var school models.School
	err := s.db.WithContext(ctx).First(&school, schoolID).Error
	if err != nil {
		return fmt.Errorf("failed to get school: %w", err)
	}

	// Update Firebase with new status
	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/packing/%s/%d", today, schoolID)
	
	updateData := map[string]interface{}{
		"school_id":   schoolID,
		"school_name": school.Name,
		"status":      status,
		"updated_at":  time.Now().Unix(),
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, updateData)
	if err != nil {
		return fmt.Errorf("failed to update Firebase: %w", err)
	}

	// If all schools are ready, send notification to logistics team
	if status == "ready" {
		err = s.checkAllSchoolsReady(ctx)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to check all schools ready: %v\n", err)
		}
	}

	return nil
}

// checkAllSchoolsReady checks if all schools for today are ready and sends notification
func (s *PackingAllocationService) checkAllSchoolsReady(ctx context.Context) error {
	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/packing/%s", today)
	
	var packingData map[string]interface{}
	err := s.dbClient.NewRef(firebasePath).Get(ctx, &packingData)
	if err != nil {
		return fmt.Errorf("failed to get packing data from Firebase: %w", err)
	}

	// Check if all schools have status "ready"
	allReady := true
	for _, data := range packingData {
		if schoolData, ok := data.(map[string]interface{}); ok {
			if status, ok := schoolData["status"].(string); ok && status != "ready" {
				allReady = false
				break
			}
		}
	}

	if allReady {
		// Send notification to logistics team
		notificationPath := "/notifications/logistics/packing_complete"
		notificationData := map[string]interface{}{
			"message":    "Semua sekolah siap untuk pengiriman",
			"date":       today,
			"timestamp":  time.Now().Unix(),
		}
		_, err = s.dbClient.NewRef(notificationPath).Push(ctx, notificationData)
		if err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil
}

// SyncPackingAllocationsToFirebase syncs packing allocations to Firebase for real-time display
func (s *PackingAllocationService) SyncPackingAllocationsToFirebase(ctx context.Context) error {
	allocations, err := s.CalculatePackingAllocations(ctx)
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	firebasePath := fmt.Sprintf("/kds/packing/%s", today)

	// Convert to map for Firebase
	firebaseData := make(map[string]interface{})
	for _, allocation := range allocations {
		firebaseData[fmt.Sprintf("%d", allocation.SchoolID)] = map[string]interface{}{
			"school_id":   allocation.SchoolID,
			"school_name": allocation.SchoolName,
			"portions":    allocation.Portions,
			"menu_items":  allocation.MenuItems,
			"status":      allocation.Status,
		}
	}

	err = s.dbClient.NewRef(firebasePath).Set(ctx, firebaseData)
	if err != nil {
		return fmt.Errorf("failed to sync to Firebase: %w", err)
	}

	return nil
}
