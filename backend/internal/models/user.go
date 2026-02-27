package models

import (
	"time"
)

// User represents a system user with authentication and role information
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	NIK          string    `gorm:"uniqueIndex;size:20;not null" json:"nik" validate:"required"`
	Email        string    `gorm:"uniqueIndex;size:100;not null" json:"email" validate:"required,email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	FullName     string    `gorm:"size:100;not null" json:"full_name" validate:"required"`
	PhoneNumber  string    `gorm:"size:20" json:"phone_number"`
	Role         string    `gorm:"size:50;not null;index" json:"role" validate:"required,oneof=kepala_sppg kepala_yayasan akuntan ahli_gizi pengadaan chef packing driver asisten_lapangan kebersihan"`
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AuditTrail records all user actions for accountability
type AuditTrail struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Timestamp time.Time `gorm:"index;not null" json:"timestamp"`
	Action    string    `gorm:"size:50;not null;index" json:"action"` // create, update, delete, login, etc.
	Entity    string    `gorm:"size:100;not null;index" json:"entity"` // table/resource name
	EntityID  string    `gorm:"size:100" json:"entity_id"`
	OldValue  string    `gorm:"type:text" json:"old_value"`
	NewValue  string    `gorm:"type:text" json:"new_value"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
