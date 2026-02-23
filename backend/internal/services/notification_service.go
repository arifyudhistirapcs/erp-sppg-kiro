package services

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/erp-sppg/backend/internal/models"
	"gorm.io/gorm"
)

// NotificationService handles notification operations
type NotificationService struct {
	db          *gorm.DB
	firebaseSync *FirebaseSyncService
}

// NotificationType constants
const (
	NotificationTypeLowStock         = "low_stock"
	NotificationTypePOApproval       = "po_approval"
	NotificationTypePackingComplete  = "packing_complete"
	NotificationTypeDeliveryComplete = "delivery_complete"
)

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, firebaseApp *firebase.App) (*NotificationService, error) {
	firebaseSync, err := NewFirebaseSyncService(firebaseApp)
	if err != nil {
		return nil, err
	}

	return &NotificationService{
		db:          db,
		firebaseSync: firebaseSync,
	}, nil
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
	notification.CreatedAt = time.Now()
	notification.IsRead = false

	if err := s.db.Create(notification).Error; err != nil {
		return fmt.Errorf("gagal membuat notifikasi: %w", err)
	}

	// Push to Firebase for real-time notification
	if err := s.pushNotificationToFirebase(ctx, notification); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Peringatan: gagal mengirim notifikasi ke Firebase: %v\n", err)
	}

	return nil
}

// SendLowStockNotification sends a low stock alert notification
func (s *NotificationService) SendLowStockNotification(ctx context.Context, userID uint, ingredientName string, currentQty, minThreshold float64) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    NotificationTypeLowStock,
		Title:   "Peringatan Stok Menipis",
		Message: fmt.Sprintf("Stok %s menipis. Jumlah saat ini: %.2f, batas minimum: %.2f", ingredientName, currentQty, minThreshold),
		Link:    "/inventory",
	}

	return s.CreateNotification(ctx, notification)
}

// SendPOApprovalNotification sends a PO approval request notification
func (s *NotificationService) SendPOApprovalNotification(ctx context.Context, userID uint, poNumber string, supplierName string, totalAmount float64) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    NotificationTypePOApproval,
		Title:   "Permintaan Persetujuan Purchase Order",
		Message: fmt.Sprintf("PO %s dari %s memerlukan persetujuan. Total: Rp %.2f", poNumber, supplierName, totalAmount),
		Link:    fmt.Sprintf("/purchase-orders/%s", poNumber),
	}

	return s.CreateNotification(ctx, notification)
}

// SendPackingCompleteNotification sends a packing complete notification
func (s *NotificationService) SendPackingCompleteNotification(ctx context.Context, userID uint, date string, totalSchools int) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    NotificationTypePackingComplete,
		Title:   "Packing Selesai",
		Message: fmt.Sprintf("Packing untuk %d sekolah pada tanggal %s telah selesai dan siap dikirim", totalSchools, date),
		Link:    "/kds/packing",
	}

	return s.CreateNotification(ctx, notification)
}

// SendDeliveryCompleteNotification sends a delivery complete notification
func (s *NotificationService) SendDeliveryCompleteNotification(ctx context.Context, userID uint, schoolName string, driverName string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    NotificationTypeDeliveryComplete,
		Title:   "Pengiriman Selesai",
		Message: fmt.Sprintf("Pengiriman ke %s oleh %s telah selesai", schoolName, driverName),
		Link:    "/delivery-tasks",
	}

	return s.CreateNotification(ctx, notification)
}

// GetUserNotifications retrieves all notifications for a user
func (s *NotificationService) GetUserNotifications(userID uint, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung notifikasi: %w", err)
	}

	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil notifikasi: %w", err)
	}

	return notifications, total, nil
}

// GetUnreadCount gets the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("gagal menghitung notifikasi yang belum dibaca: %w", err)
	}
	return count, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID, userID uint) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)

	if result.Error != nil {
		return fmt.Errorf("gagal menandai notifikasi sebagai dibaca: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notifikasi tidak ditemukan")
	}

	// Update Firebase
	path := fmt.Sprintf("/notifications/%d/%d", userID, notificationID)
	if err := s.firebaseSync.UpdateField(ctx, path, map[string]interface{}{
		"is_read": true,
	}); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Peringatan: gagal mengupdate notifikasi di Firebase: %v\n", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("gagal menandai semua notifikasi sebagai dibaca: %w", err)
	}

	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("gagal menghapus notifikasi: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notifikasi tidak ditemukan")
	}

	// Delete from Firebase
	path := fmt.Sprintf("/notifications/%d/%d", userID, notificationID)
	if err := s.firebaseSync.DeletePath(ctx, path); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Peringatan: gagal menghapus notifikasi dari Firebase: %v\n", err)
	}

	return nil
}

// pushNotificationToFirebase pushes notification to Firebase for real-time delivery
func (s *NotificationService) pushNotificationToFirebase(ctx context.Context, notification *models.Notification) error {
	path := fmt.Sprintf("/notifications/%d/%d", notification.UserID, notification.ID)
	
	data := map[string]interface{}{
		"id":         notification.ID,
		"type":       notification.Type,
		"title":      notification.Title,
		"message":    notification.Message,
		"is_read":    notification.IsRead,
		"link":       notification.Link,
		"created_at": notification.CreatedAt.Unix(),
	}

	return s.firebaseSync.PushUpdate(ctx, path, data)
}
