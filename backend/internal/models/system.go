package models

import (
	"time"
)

// SystemConfig represents system configuration parameters
type SystemConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:100;not null" json:"key" validate:"required"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	DataType  string    `gorm:"size:20;not null" json:"data_type" validate:"required,oneof=string int float bool json"` // string, int, float, bool, json
	Category  string    `gorm:"size:50;index" json:"category"`                                                          // nutrition, inventory, system, etc.
	UpdatedBy uint      `gorm:"not null;index" json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
	Updater   User      `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

// Notification represents system notifications for users
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Type      string    `gorm:"size:50;not null;index" json:"type" validate:"required"` // low_stock, po_approval, delivery_complete, etc.
	Title     string    `gorm:"size:200;not null" json:"title" validate:"required"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"default:false;index" json:"is_read"`
	Link      string    `gorm:"size:500" json:"link"` // deep link to relevant screen
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
