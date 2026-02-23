package models

import (
	"time"
)

// Employee represents an employee in the organization
type Employee struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex;not null" json:"user_id"` // links to User table
	NIK         string    `gorm:"uniqueIndex;size:20;not null" json:"nik" validate:"required"`
	FullName    string    `gorm:"size:100;not null;index" json:"full_name" validate:"required"`
	Email       string    `gorm:"uniqueIndex;size:100;not null" json:"email" validate:"required,email"`
	PhoneNumber string    `gorm:"size:20" json:"phone_number"`
	Position    string    `gorm:"size:100;index" json:"position"`
	JoinDate    time.Time `gorm:"not null" json:"join_date"`
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Attendance represents employee attendance records
type Attendance struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	EmployeeID uint       `gorm:"index;not null" json:"employee_id"`
	Date       time.Time  `gorm:"index;not null" json:"date"`
	CheckIn    time.Time  `gorm:"not null" json:"check_in"`
	CheckOut   *time.Time `json:"check_out"`
	WorkHours  float64    `gorm:"default:0" json:"work_hours"`
	SSID       string     `gorm:"size:100" json:"ssid"`
	BSSID      string     `gorm:"size:100" json:"bssid"`
	CreatedAt  time.Time  `json:"created_at"`
	Employee   Employee   `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// WiFiConfig represents authorized Wi-Fi networks for attendance
type WiFiConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SSID      string    `gorm:"size:100;not null;index" json:"ssid" validate:"required"`
	BSSID     string    `gorm:"size:100;not null;index" json:"bssid" validate:"required"`
	Location  string    `gorm:"size:200" json:"location"`
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
