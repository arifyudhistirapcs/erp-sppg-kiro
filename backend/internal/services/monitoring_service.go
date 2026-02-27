package services

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"gorm.io/gorm"

	"github.com/erp-sppg/backend/internal/models"
)

// syncRetryItem represents an item in the retry queue
type syncRetryItem struct {
	recordID    uint
	attempt     int
	nextRetryAt time.Time
}

// MonitoringService handles logistics monitoring operations
type MonitoringService struct {
	db          *gorm.DB
	firebaseApp *firebase.App
	dbClient    *db.Client
	retryQueue  chan syncRetryItem
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(database *gorm.DB, firebaseApp *firebase.App) (*MonitoringService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase database client: %w", err)
	}

	service := &MonitoringService{
		db:          database,
		firebaseApp: firebaseApp,
		dbClient:    dbClient,
		retryQueue:  make(chan syncRetryItem, 100), // Buffer for 100 retry items
	}

	// Start retry worker goroutine
	go service.retryWorker()

	return service, nil
}

// GetDeliveryRecords retrieves delivery records for a specific date with optional filters
func (s *MonitoringService) GetDeliveryRecords(date time.Time, filters map[string]interface{}) ([]models.DeliveryRecord, error) {
	var records []models.DeliveryRecord

	// Start with base query filtering by date
	query := s.db.Where("DATE(delivery_date) = DATE(?)", date)

	// Apply optional filters
	if schoolID, ok := filters["school_id"]; ok {
		query = query.Where("school_id = ?", schoolID)
	}

	if status, ok := filters["status"]; ok {
		query = query.Where("current_status = ?", status)
	}

	if driverID, ok := filters["driver_id"]; ok {
		query = query.Where("driver_id = ?", driverID)
	}

	// Preload associations
	err := query.
		Preload("School").
		Preload("Driver").
		Preload("MenuItem").
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetDeliveryRecordDetail retrieves a single delivery record by ID with all associations
func (s *MonitoringService) GetDeliveryRecordDetail(recordID uint) (*models.DeliveryRecord, error) {
	var record models.DeliveryRecord

	// Query delivery record by ID and preload all associations
	err := s.db.
		Preload("School").
		Preload("Driver").
		Preload("MenuItem").
		First(&record, recordID).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// UpdateDeliveryStatus updates the status of a delivery record with validation
// and creates a status transition record. It performs the following steps:
// 1. Retrieves the current delivery record
// 2. Validates the status transition using ValidateStatusTransition
// 3. Validates the stage sequence using ValidateStageSequence
// 4. Updates the delivery record's current_status in a database transaction
// 5. Creates a StatusTransition record with timestamp and user attribution
// 6. Triggers Firebase sync asynchronously (non-blocking)
//
// Parameters:
//   - recordID: The ID of the delivery record to update
//   - newStatus: The new status to transition to
//   - userID: The ID of the user performing the status update
//   - notes: Optional notes about the status transition
//
// Returns an error if:
//   - The delivery record is not found
//   - The status transition is invalid (violates transition rules)
//   - The stage sequence is invalid (e.g., trying to skip stages)
//   - Database transaction fails
//
// Requirements: 2.1-2.8, 3.1-3.5, 9.1, 13.1-13.5
func (s *MonitoringService) UpdateDeliveryStatus(recordID uint, newStatus string, userID uint, notes string) error {
	// Step 1: Retrieve the current delivery record
	var record models.DeliveryRecord
	if err := s.db.First(&record, recordID).Error; err != nil {
		return err
	}

	currentStatus := record.CurrentStatus

	// Step 2: Validate status transition
	if err := ValidateStatusTransition(currentStatus, newStatus); err != nil {
		return err
	}

	// Step 3: Validate stage sequence
	if err := ValidateStageSequence(currentStatus, newStatus); err != nil {
		return err
	}

	// Step 4 & 5: Update delivery record and create status transition in a transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Update delivery record's current_status
		if err := tx.Model(&record).Update("current_status", newStatus).Error; err != nil {
			return err
		}

		// Create StatusTransition record
		transition := models.StatusTransition{
			DeliveryRecordID: recordID,
			FromStatus:       currentStatus,
			ToStatus:         newStatus,
			TransitionedAt:   time.Now(),
			TransitionedBy:   userID,
			Notes:            notes,
		}

		if err := tx.Create(&transition).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Step 6: Trigger Firebase sync asynchronously (non-blocking)
	// Update the record's current status for Firebase sync
	record.CurrentStatus = newStatus

	// Only spawn goroutine if Firebase is configured
	if s.dbClient != nil {
		go func() {
			if syncErr := s.syncToFirebaseWithRetry(&record, 0); syncErr != nil {
				// Log error but don't block the main operation
				log.Printf("Firebase sync failed for delivery record %d after all retries: %v", record.ID, syncErr)
			}
		}()
	}

	return nil
}

// GetActivityLog retrieves the activity log (status transition history) for a delivery record.
// It performs the following steps:
// 1. Queries status_transitions table where delivery_record_id = recordID
// 2. Orders results by transitioned_at in chronological order (ASC)
// 3. Preloads User association for each transition to get transitioned_by user details
// 4. Returns the complete activity log with all transitions
//
// The elapsed time between consecutive transitions should be calculated by the caller
// using the transitioned_at timestamps.
//
// Parameters:
//   - recordID: The ID of the delivery record to get activity log for
//
// Returns:
//   - []models.StatusTransition: Array of status transitions in chronological order
//   - error: Error if query fails or record not found
//
// Requirements: 1.5, 9.2, 9.3
func (s *MonitoringService) GetActivityLog(recordID uint) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition

	// Query status_transitions for the delivery record
	// Order by transitioned_at chronologically (ASC)
	// Preload User association for transitioned_by
	err := s.db.
		Where("delivery_record_id = ?", recordID).
		Order("transitioned_at ASC").
		Preload("User").
		Find(&transitions).Error

	if err != nil {
		return nil, err
	}

	return transitions, nil
}

// GetDailySummary retrieves summary statistics for deliveries on a specific date.
// It performs the following calculations:
// 1. Counts total delivery records for the date
// 2. Counts records grouped by current_status
// 3. Counts completed deliveries (status = "sudah_diterima_pihak_sekolah")
// 4. Counts ompreng in cleaning (status = "ompreng_proses_pencucian")
// 5. Counts ompreng cleaned (status = "ompreng_selesai_dicuci")
//
// Parameters:
//   - date: The date to get summary statistics for
//
// Returns:
//   - *models.DailySummary: Summary statistics with all counts
//   - error: Error if query fails
//
// Requirements: 15.1, 15.2, 15.3, 15.4, 15.5
func (s *MonitoringService) GetDailySummary(date time.Time) (*models.DailySummary, error) {
	var summary models.DailySummary

	// Initialize status counts map
	summary.StatusCounts = make(map[string]int)

	// Count total delivery records for the date
	var totalCount int64
	if err := s.db.Model(&models.DeliveryRecord{}).
		Where("DATE(delivery_date) = DATE(?)", date).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	summary.TotalDeliveries = int(totalCount)

	// Count records by each status
	var statusResults []struct {
		CurrentStatus string
		Count         int64
	}
	if err := s.db.Model(&models.DeliveryRecord{}).
		Select("current_status, COUNT(*) as count").
		Where("DATE(delivery_date) = DATE(?)", date).
		Group("current_status").
		Scan(&statusResults).Error; err != nil {
		return nil, err
	}

	// Populate status counts map
	for _, result := range statusResults {
		summary.StatusCounts[result.CurrentStatus] = int(result.Count)
	}

	// Count completed deliveries (status = "sudah_diterima_pihak_sekolah")
	var completedCount int64
	if err := s.db.Model(&models.DeliveryRecord{}).
		Where("DATE(delivery_date) = DATE(?) AND current_status = ?", date, "sudah_diterima_pihak_sekolah").
		Count(&completedCount).Error; err != nil {
		return nil, err
	}
	summary.CompletedDeliveries = int(completedCount)

	// Count ompreng in cleaning (status = "ompreng_proses_pencucian")
	var cleaningCount int64
	if err := s.db.Model(&models.DeliveryRecord{}).
		Where("DATE(delivery_date) = DATE(?) AND current_status = ?", date, "ompreng_proses_pencucian").
		Count(&cleaningCount).Error; err != nil {
		return nil, err
	}
	summary.OmprengInCleaning = int(cleaningCount)

	// Count ompreng cleaned (status = "ompreng_selesai_dicuci")
	var cleanedCount int64
	if err := s.db.Model(&models.DeliveryRecord{}).
		Where("DATE(delivery_date) = DATE(?) AND current_status = ?", date, "ompreng_selesai_dicuci").
		Count(&cleanedCount).Error; err != nil {
		return nil, err
	}
	summary.OmprengCleaned = int(cleanedCount)

	return &summary, nil
}

// syncToFirebase synchronizes delivery record data to Firebase Realtime Database
// for real-time updates across all connected clients. The data is written to
// /monitoring/deliveries/{date}/record_{id} path.
//
// This method formats the delivery record data and writes it to Firebase.
// Errors are returned but should be handled gracefully by the caller,
// typically with retry logic.
//
// Requirements: 1.1
func (s *MonitoringService) syncToFirebase(record *models.DeliveryRecord) error {
	// Format date as YYYY-MM-DD for Firebase path
	dateStr := record.DeliveryDate.Format("2006-01-02")

	// Construct Firebase path: /monitoring/deliveries/{date}/record_{id}
	firebasePath := fmt.Sprintf("/monitoring/deliveries/%s/record_%d", dateStr, record.ID)

	// Prepare data for Firebase
	// Note: We need to preload associations if they're not already loaded
	var fullRecord models.DeliveryRecord
	if err := s.db.
		Preload("School").
		Preload("Driver").
		First(&fullRecord, record.ID).Error; err != nil {
		return err
	}

	// Format data for Firebase
	data := map[string]interface{}{
		"id":             fullRecord.ID,
		"school_name":    fullRecord.School.Name,
		"driver_name":    fullRecord.Driver.FullName,
		"current_status": fullRecord.CurrentStatus,
		"portions":       fullRecord.Portions,
		"ompreng_count":  fullRecord.OmprengCount,
		"last_updated":   time.Now().Unix(),
	}

	// Write to Firebase
	if s.dbClient != nil {
		ref := s.dbClient.NewRef(firebasePath)
		ctx := context.Background()
		if err := ref.Set(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// syncToFirebaseWithRetry attempts to sync to Firebase with exponential backoff retry logic.
// It implements the following retry strategy:
// - Attempt 1: Immediate
// - Attempt 2: 1 second delay
// - Attempt 3: 2 seconds delay
// - Attempt 4: 4 seconds delay
// - Attempt 5: 8 seconds delay
// - Attempt 6: 16 seconds delay
// - Maximum retry attempts: 5 (total 6 attempts including initial)
//
// After max retries, the error is logged for admin alerting.
//
// Parameters:
//   - record: The delivery record to sync
//   - attempt: Current attempt number (0-based)
//
// Returns:
//   - error: Error if all retry attempts fail
//
// Requirements: 1.1
func (s *MonitoringService) syncToFirebaseWithRetry(record *models.DeliveryRecord, attempt int) error {
	const maxRetries = 5

	// Attempt to sync to Firebase
	err := s.syncToFirebase(record)

	if err == nil {
		// Success
		return nil
	}

	// Check if we've exceeded max retries
	if attempt >= maxRetries {
		// Log error for admin alerting
		log.Printf("ERROR: Firebase sync failed for delivery record %d after %d attempts: %v",
			record.ID, attempt+1, err)
		return fmt.Errorf("firebase sync failed after %d attempts: %w", attempt+1, err)
	}

	// Calculate exponential backoff delay: 1s, 2s, 4s, 8s, 16s
	backoffSeconds := 1 << attempt // 2^attempt
	delay := time.Duration(backoffSeconds) * time.Second

	log.Printf("Firebase sync failed for delivery record %d (attempt %d/%d), retrying in %v: %v",
		record.ID, attempt+1, maxRetries+1, delay, err)

	// Queue for retry with delay
	go func() {
		time.Sleep(delay)
		if retryErr := s.syncToFirebaseWithRetry(record, attempt+1); retryErr != nil {
			// Final error already logged in the recursive call
			_ = retryErr
		}
	}()

	return nil
}

// retryWorker processes items from the retry queue
// This is a background worker that handles delayed retries
func (s *MonitoringService) retryWorker() {
	for item := range s.retryQueue {
		// Wait until it's time to retry
		now := time.Now()
		if item.nextRetryAt.After(now) {
			time.Sleep(item.nextRetryAt.Sub(now))
		}

		// Retrieve the delivery record
		var record models.DeliveryRecord
		if err := s.db.First(&record, item.recordID).Error; err != nil {
			log.Printf("Failed to retrieve delivery record %d for retry: %v", item.recordID, err)
			continue
		}

		// Attempt sync with retry
		go func() {
			if err := s.syncToFirebaseWithRetry(&record, item.attempt); err != nil {
				log.Printf("Retry failed for delivery record %d: %v", record.ID, err)
			}
		}()
	}
}

// queueForRetry adds a delivery record to the retry queue
func (s *MonitoringService) queueForRetry(recordID uint, attempt int) {
	// Calculate next retry time with exponential backoff
	backoffSeconds := 1 << attempt // 2^attempt
	nextRetryAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

	item := syncRetryItem{
		recordID:    recordID,
		attempt:     attempt,
		nextRetryAt: nextRetryAt,
	}

	// Non-blocking send to queue
	select {
	case s.retryQueue <- item:
		log.Printf("Queued delivery record %d for retry (attempt %d) at %v",
			recordID, attempt+1, nextRetryAt)
	default:
		log.Printf("WARNING: Retry queue full, dropping retry for delivery record %d", recordID)
	}
}
