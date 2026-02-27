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

// cleaningSyncRetryItem represents an item in the cleaning retry queue
type cleaningSyncRetryItem struct {
	cleaningID  uint
	attempt     int
	nextRetryAt time.Time
}

// CleaningService handles ompreng cleaning operations
type CleaningService struct {
	db          *gorm.DB
	firebaseApp *firebase.App
	dbClient    *db.Client
	retryQueue  chan cleaningSyncRetryItem
}

// NewCleaningService creates a new cleaning service
func NewCleaningService(database *gorm.DB, firebaseApp *firebase.App) (*CleaningService, error) {
	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase database client: %w", err)
	}

	service := &CleaningService{
		db:          database,
		firebaseApp: firebaseApp,
		dbClient:    dbClient,
		retryQueue:  make(chan cleaningSyncRetryItem, 100), // Buffer for 100 retry items
	}

	// Start retry worker goroutine
	go service.retryWorker()

	return service, nil
}

// GetPendingOmpreng retrieves ompreng cleaning records that are pending cleaning.
// It queries for:
// 1. ompreng_cleanings with status "pending", OR
// 2. delivery_records with status "ompreng_sampai_di_sppg"
//
// The method preloads DeliveryRecord with School association to provide
// complete information for the KDS Cleaning interface.
//
// Returns:
//   - []models.OmprengCleaning: Array of pending ompreng cleaning records
//   - error: Error if query fails
//
// Requirements: 7.1, 7.4
func (s *CleaningService) GetPendingOmpreng(dateStr string) ([]models.OmprengCleaning, error) {
	var cleanings []models.OmprengCleaning

	// Build query with optional date filter
	query := s.db.Preload("DeliveryRecord.School").
		Where("cleaning_status = ?", "pending")

	// Add date filter if provided
	if dateStr != "" {
		// Parse date
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			// Filter by delivery_date in delivery_records
			query = query.Joins("JOIN delivery_records ON delivery_records.id = ompreng_cleanings.delivery_record_id").
				Where("DATE(delivery_records.delivery_date) = DATE(?)", date)
		}
	}

	err := query.Find(&cleanings).Error
	if err != nil {
		return nil, err
	}

	// Also check for delivery records with status "ompreng_sampai_di_sppg"
	// that don't have a cleaning record yet
	deliveryQuery := s.db.Preload("School").
		Where("current_status = ?", "ompreng_sampai_di_sppg")

	// Add date filter for delivery records
	if dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			deliveryQuery = deliveryQuery.Where("DATE(delivery_date) = DATE(?)", date)
		}
	}

	var deliveryRecords []models.DeliveryRecord
	err = deliveryQuery.Find(&deliveryRecords).Error
	if err != nil {
		return nil, err
	}

	// Create cleaning records for delivery records that don't have one
	for _, record := range deliveryRecords {
		// Check if cleaning record already exists
		var existingCleaning models.OmprengCleaning
		err := s.db.Where("delivery_record_id = ?", record.ID).First(&existingCleaning).Error

		if err == gorm.ErrRecordNotFound {
			// Create new cleaning record
			cleaning := models.OmprengCleaning{
				DeliveryRecordID: record.ID,
				OmprengCount:     record.OmprengCount,
				CleaningStatus:   "pending",
				DeliveryRecord:   record,
			}
			cleanings = append(cleanings, cleaning)
		}
	}

	return cleanings, nil
}

// StartCleaning updates an ompreng cleaning record to start the cleaning process.
// It performs the following steps:
// 1. Updates ompreng_cleaning status to "in_progress"
// 2. Sets started_at timestamp to current time
// 3. Sets cleaned_by to current user ID
// 4. Updates corresponding delivery record status to "ompreng_proses_pencucian"
// 5. Syncs to Firebase asynchronously
//
// Parameters:
//   - cleaningID: The ID of the ompreng cleaning record
//   - userID: The ID of the cleaning staff member starting the cleaning
//
// Returns:
//   - error: Error if update fails or record not found
//
// Requirements: 4.1, 7.2, 7.5
func (s *CleaningService) StartCleaning(cleaningID uint, userID uint) error {
	// Retrieve the cleaning record
	var cleaning models.OmprengCleaning
	if err := s.db.First(&cleaning, cleaningID).Error; err != nil {
		return err
	}

	// Update cleaning record and delivery record in a transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Update ompreng_cleaning status to "in_progress"
		now := time.Now()
		updates := map[string]interface{}{
			"cleaning_status": "in_progress",
			"started_at":      now,
			"cleaned_by":      userID,
		}

		if err := tx.Model(&cleaning).Updates(updates).Error; err != nil {
			return err
		}

		// Update corresponding delivery record status to "ompreng_proses_pencucian"
		if err := tx.Model(&models.DeliveryRecord{}).
			Where("id = ?", cleaning.DeliveryRecordID).
			Update("current_status", "ompreng_proses_pencucian").Error; err != nil {
			return err
		}

		// Create status transition record
		transition := models.StatusTransition{
			DeliveryRecordID: cleaning.DeliveryRecordID,
			FromStatus:       "ompreng_sampai_di_sppg",
			ToStatus:         "ompreng_proses_pencucian",
			TransitionedAt:   now,
			TransitionedBy:   userID,
			Notes:            "Cleaning started",
		}

		if err := tx.Create(&transition).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Update the cleaning record with new values for Firebase sync
	cleaning.CleaningStatus = "in_progress"
	now := time.Now()
	cleaning.StartedAt = &now
	cleaning.CleanedBy = &userID

	// Sync to Firebase asynchronously with retry mechanism
	// Only spawn goroutine if Firebase is configured
	if s.dbClient != nil {
		go func() {
			if syncErr := s.syncToFirebaseWithRetry(&cleaning, 0); syncErr != nil {
				log.Printf("Firebase sync failed for cleaning record %d after all retries: %v", cleaning.ID, syncErr)
			}
		}()
	}

	return nil
}

// CompleteCleaning updates an ompreng cleaning record to mark cleaning as completed.
// It performs the following steps:
// 1. Updates ompreng_cleaning status to "completed"
// 2. Sets completed_at timestamp to current time
// 3. Updates corresponding delivery record status to "ompreng_selesai_dicuci"
// 4. Syncs to Firebase asynchronously
//
// Parameters:
//   - cleaningID: The ID of the ompreng cleaning record
//   - userID: The ID of the user completing the cleaning (for audit trail)
//
// Returns:
//   - error: Error if update fails or record not found
//
// Requirements: 4.2, 7.3, 7.5
func (s *CleaningService) CompleteCleaning(cleaningID uint, userID uint) error {
	// Retrieve the cleaning record
	var cleaning models.OmprengCleaning
	if err := s.db.First(&cleaning, cleaningID).Error; err != nil {
		return err
	}

	// Update cleaning record and delivery record in a transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Update ompreng_cleaning status to "completed"
		now := time.Now()
		updates := map[string]interface{}{
			"cleaning_status": "completed",
			"completed_at":    now,
		}

		if err := tx.Model(&cleaning).Updates(updates).Error; err != nil {
			return err
		}

		// Update corresponding delivery record status to "ompreng_selesai_dicuci"
		if err := tx.Model(&models.DeliveryRecord{}).
			Where("id = ?", cleaning.DeliveryRecordID).
			Update("current_status", "ompreng_selesai_dicuci").Error; err != nil {
			return err
		}

		// Create status transition record
		transition := models.StatusTransition{
			DeliveryRecordID: cleaning.DeliveryRecordID,
			FromStatus:       "ompreng_proses_pencucian",
			ToStatus:         "ompreng_selesai_dicuci",
			TransitionedAt:   now,
			TransitionedBy:   userID,
			Notes:            "Cleaning completed",
		}

		if err := tx.Create(&transition).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Update the cleaning record with new values for Firebase sync
	cleaning.CleaningStatus = "completed"
	now := time.Now()
	cleaning.CompletedAt = &now

	// Sync to Firebase asynchronously with retry mechanism
	// Only spawn goroutine if Firebase is configured
	if s.dbClient != nil {
		go func() {
			if syncErr := s.syncToFirebaseWithRetry(&cleaning, 0); syncErr != nil {
				log.Printf("Firebase sync failed for cleaning record %d after all retries: %v", cleaning.ID, syncErr)
			}
		}()
	}

	return nil
}

// SyncToFirebase synchronizes cleaning record data to Firebase Realtime Database
// for real-time updates in the KDS Cleaning interface. The data is written to
// /cleaning/pending/{cleaning_id} path.
//
// This method formats the cleaning record data and writes it to Firebase.
// Errors are returned but should be handled gracefully by the caller,
// typically with retry logic.
//
// Parameters:
//   - cleaning: The ompreng cleaning record to sync
//
// Returns:
//   - error: Error if Firebase write fails
//
// Requirements: 7.1
func (s *CleaningService) SyncToFirebase(cleaning *models.OmprengCleaning) error {
	// Construct Firebase path: /cleaning/pending/{cleaning_id}
	firebasePath := fmt.Sprintf("/cleaning/pending/cleaning_%d", cleaning.ID)

	// Prepare data for Firebase
	// Preload associations if not already loaded
	var fullCleaning models.OmprengCleaning
	if err := s.db.
		Preload("DeliveryRecord.School").
		First(&fullCleaning, cleaning.ID).Error; err != nil {
		return err
	}

	// Format data for Firebase
	data := map[string]interface{}{
		"id":                 fullCleaning.ID,
		"delivery_record_id": fullCleaning.DeliveryRecordID,
		"ompreng_count":      fullCleaning.OmprengCount,
		"status":             fullCleaning.CleaningStatus,
	}

	// Add school name if available
	if fullCleaning.DeliveryRecord.School.Name != "" {
		data["school_name"] = fullCleaning.DeliveryRecord.School.Name
	}

	// Add timestamps if available
	if fullCleaning.StartedAt != nil {
		data["started_at"] = fullCleaning.StartedAt.Unix()
	}
	if fullCleaning.CompletedAt != nil {
		data["completed_at"] = fullCleaning.CompletedAt.Unix()
	}
	if fullCleaning.DeliveryRecord.DeliveryDate.IsZero() == false {
		data["delivery_date"] = fullCleaning.DeliveryRecord.DeliveryDate.Format("2006-01-02")
	}

	// Add last updated timestamp
	data["last_updated"] = time.Now().Unix()

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
//   - cleaning: The ompreng cleaning record to sync
//   - attempt: Current attempt number (0-based)
//
// Returns:
//   - error: Error if all retry attempts fail
//
// Requirements: 1.1
func (s *CleaningService) syncToFirebaseWithRetry(cleaning *models.OmprengCleaning, attempt int) error {
	const maxRetries = 5

	// Attempt to sync to Firebase
	err := s.SyncToFirebase(cleaning)

	if err == nil {
		// Success
		return nil
	}

	// Check if we've exceeded max retries
	if attempt >= maxRetries {
		// Log error for admin alerting
		log.Printf("ERROR: Firebase sync failed for cleaning record %d after %d attempts: %v",
			cleaning.ID, attempt+1, err)
		return fmt.Errorf("firebase sync failed after %d attempts: %w", attempt+1, err)
	}

	// Calculate exponential backoff delay: 1s, 2s, 4s, 8s, 16s
	backoffSeconds := 1 << attempt // 2^attempt
	delay := time.Duration(backoffSeconds) * time.Second

	log.Printf("Firebase sync failed for cleaning record %d (attempt %d/%d), retrying in %v: %v",
		cleaning.ID, attempt+1, maxRetries+1, delay, err)

	// Queue for retry with delay
	go func() {
		time.Sleep(delay)
		if retryErr := s.syncToFirebaseWithRetry(cleaning, attempt+1); retryErr != nil {
			// Final error already logged in the recursive call
			_ = retryErr
		}
	}()

	return nil
}

// retryWorker processes items from the retry queue
// This is a background worker that handles delayed retries
func (s *CleaningService) retryWorker() {
	for item := range s.retryQueue {
		// Wait until it's time to retry
		now := time.Now()
		if item.nextRetryAt.After(now) {
			time.Sleep(item.nextRetryAt.Sub(now))
		}

		// Retrieve the cleaning record
		var cleaning models.OmprengCleaning
		if err := s.db.First(&cleaning, item.cleaningID).Error; err != nil {
			log.Printf("Failed to retrieve cleaning record %d for retry: %v", item.cleaningID, err)
			continue
		}

		// Attempt sync with retry
		go func() {
			if err := s.syncToFirebaseWithRetry(&cleaning, item.attempt); err != nil {
				log.Printf("Retry failed for cleaning record %d: %v", cleaning.ID, err)
			}
		}()
	}
}

// queueForRetry adds a cleaning record to the retry queue
func (s *CleaningService) queueForRetry(cleaningID uint, attempt int) {
	// Calculate next retry time with exponential backoff
	backoffSeconds := 1 << attempt // 2^attempt
	nextRetryAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

	item := cleaningSyncRetryItem{
		cleaningID:  cleaningID,
		attempt:     attempt,
		nextRetryAt: nextRetryAt,
	}

	// Non-blocking send to queue
	select {
	case s.retryQueue <- item:
		log.Printf("Queued cleaning record %d for retry (attempt %d) at %v",
			cleaningID, attempt+1, nextRetryAt)
	default:
		log.Printf("WARNING: Retry queue full, dropping retry for cleaning record %d", cleaningID)
	}
}
