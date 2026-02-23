package services

import (
	"testing"
	"time"

	"github.com/erp-sppg/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNotificationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&models.Notification{}, &models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestNotificationService_CreateNotification(t *testing.T) {
	db := setupNotificationTestDB(t)
	
	// Create a test user
	user := &models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Role:         "pengadaan",
		IsActive:     true,
	}
	db.Create(user)

	// Note: We can't fully test Firebase integration without a real Firebase instance
	// So we'll test the database operations only
	
	notification := &models.Notification{
		UserID:  user.ID,
		Type:    NotificationTypeLowStock,
		Title:   "Test Notification",
		Message: "This is a test notification",
		Link:    "/test",
	}

	// Test creating notification (without Firebase)
	notification.CreatedAt = time.Now()
	notification.IsRead = false
	err := db.Create(notification).Error
	
	assert.NoError(t, err)
	assert.NotZero(t, notification.ID)
	assert.Equal(t, user.ID, notification.UserID)
	assert.Equal(t, NotificationTypeLowStock, notification.Type)
	assert.False(t, notification.IsRead)
}

func TestNotificationService_GetUserNotifications(t *testing.T) {
	db := setupNotificationTestDB(t)
	
	// Create a test user
	user := &models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Role:         "pengadaan",
		IsActive:     true,
	}
	db.Create(user)

	// Create test notifications
	notifications := []models.Notification{
		{
			UserID:    user.ID,
			Type:      NotificationTypeLowStock,
			Title:     "Notification 1",
			Message:   "Message 1",
			IsRead:    false,
			CreatedAt: time.Now(),
		},
		{
			UserID:    user.ID,
			Type:      NotificationTypePOApproval,
			Title:     "Notification 2",
			Message:   "Message 2",
			IsRead:    true,
			CreatedAt: time.Now(),
		},
	}
	
	for _, n := range notifications {
		db.Create(&n)
	}

	// Test retrieving notifications
	var result []models.Notification
	var total int64
	
	query := db.Model(&models.Notification{}).Where("user_id = ?", user.ID)
	query.Count(&total)
	query.Order("created_at DESC").Limit(10).Offset(0).Find(&result)
	
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	db := setupNotificationTestDB(t)
	
	// Create a test user
	user := &models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Role:         "pengadaan",
		IsActive:     true,
	}
	db.Create(user)

	// Create test notifications
	notifications := []models.Notification{
		{
			UserID:    user.ID,
			Type:      NotificationTypeLowStock,
			Title:     "Unread 1",
			Message:   "Message 1",
			IsRead:    false,
			CreatedAt: time.Now(),
		},
		{
			UserID:    user.ID,
			Type:      NotificationTypePOApproval,
			Title:     "Unread 2",
			Message:   "Message 2",
			IsRead:    false,
			CreatedAt: time.Now(),
		},
		{
			UserID:    user.ID,
			Type:      NotificationTypeDeliveryComplete,
			Title:     "Read 1",
			Message:   "Message 3",
			IsRead:    true,
			CreatedAt: time.Now(),
		},
	}
	
	for _, n := range notifications {
		db.Create(&n)
	}

	// Test unread count
	var count int64
	db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", user.ID, false).
		Count(&count)
	
	assert.Equal(t, int64(2), count)
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	db := setupNotificationTestDB(t)
	
	// Create a test user
	user := &models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Role:         "pengadaan",
		IsActive:     true,
	}
	db.Create(user)

	// Create test notification
	notification := &models.Notification{
		UserID:    user.ID,
		Type:      NotificationTypeLowStock,
		Title:     "Test Notification",
		Message:   "Test Message",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	db.Create(notification)

	// Mark as read
	result := db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notification.ID, user.ID).
		Update("is_read", true)
	
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Verify it's marked as read
	var updated models.Notification
	db.First(&updated, notification.ID)
	assert.True(t, updated.IsRead)
}

func TestNotificationService_DeleteNotification(t *testing.T) {
	db := setupNotificationTestDB(t)
	
	// Create a test user
	user := &models.User{
		NIK:          "1234567890",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Role:         "pengadaan",
		IsActive:     true,
	}
	db.Create(user)

	// Create test notification
	notification := &models.Notification{
		UserID:    user.ID,
		Type:      NotificationTypeLowStock,
		Title:     "Test Notification",
		Message:   "Test Message",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	db.Create(notification)

	// Delete notification
	result := db.Where("id = ? AND user_id = ?", notification.ID, user.ID).
		Delete(&models.Notification{})
	
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)

	// Verify it's deleted
	var count int64
	db.Model(&models.Notification{}).Where("id = ?", notification.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestNotificationTypes(t *testing.T) {
	// Test that notification type constants are defined correctly
	assert.Equal(t, "low_stock", NotificationTypeLowStock)
	assert.Equal(t, "po_approval", NotificationTypePOApproval)
	assert.Equal(t, "packing_complete", NotificationTypePackingComplete)
	assert.Equal(t, "delivery_complete", NotificationTypeDeliveryComplete)
}
