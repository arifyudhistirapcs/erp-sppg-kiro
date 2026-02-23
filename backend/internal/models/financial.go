package models

import (
	"time"
)

// KitchenAsset represents kitchen equipment and assets
type KitchenAsset struct {
	ID                 uint                `gorm:"primaryKey" json:"id"`
	AssetCode          string              `gorm:"uniqueIndex;size:50;not null" json:"asset_code" validate:"required"`
	Name               string              `gorm:"size:200;not null;index" json:"name" validate:"required"`
	Category           string              `gorm:"size:50;index" json:"category"`
	PurchaseDate       time.Time           `gorm:"index;not null" json:"purchase_date"`
	PurchasePrice      float64             `gorm:"not null" json:"purchase_price" validate:"required,gte=0"`
	CurrentValue       float64             `gorm:"not null" json:"current_value"`
	DepreciationRate   float64             `gorm:"not null" json:"depreciation_rate" validate:"gte=0,lte=100"` // annual percentage
	Condition          string              `gorm:"size:50;index" json:"condition" validate:"oneof=good fair poor"`
	Location           string              `gorm:"size:100" json:"location"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	MaintenanceRecords []AssetMaintenance  `gorm:"foreignKey:AssetID" json:"maintenance_records,omitempty"`
}

// AssetMaintenance represents maintenance activities for assets
type AssetMaintenance struct {
	ID              uint         `gorm:"primaryKey" json:"id"`
	AssetID         uint         `gorm:"index;not null" json:"asset_id"`
	MaintenanceDate time.Time    `gorm:"index;not null" json:"maintenance_date"`
	Description     string       `gorm:"type:text" json:"description"`
	Cost            float64      `gorm:"not null" json:"cost" validate:"gte=0"`
	PerformedBy     string       `gorm:"size:100" json:"performed_by"`
	CreatedAt       time.Time    `json:"created_at"`
	Asset           KitchenAsset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// CashFlowEntry represents a financial transaction
type CashFlowEntry struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID string    `gorm:"uniqueIndex;size:50;not null" json:"transaction_id" validate:"required"`
	Date          time.Time `gorm:"index;not null" json:"date"`
	Category      string    `gorm:"size:50;not null;index" json:"category" validate:"required,oneof=bahan_baku gaji utilitas operasional lainnya"` // bahan_baku, gaji, utilitas, operasional, lainnya
	Type          string    `gorm:"size:20;not null;index" json:"type" validate:"required,oneof=income expense"`                                  // income, expense
	Amount        float64   `gorm:"not null" json:"amount" validate:"required,gt=0"`
	Description   string    `gorm:"type:text" json:"description"`
	Reference     string    `gorm:"size:100;index" json:"reference"` // GRN number, employee ID, etc.
	CreatedBy     uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	Creator       User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// BudgetTarget represents budget targets and actuals
type BudgetTarget struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Year      int       `gorm:"index;not null" json:"year" validate:"required,gte=2000"`
	Month     int       `gorm:"index;not null" json:"month" validate:"required,gte=1,lte=12"`
	Category  string    `gorm:"size:50;not null;index" json:"category" validate:"required"`
	Target    float64   `gorm:"not null" json:"target" validate:"gte=0"`
	Actual    float64   `gorm:"default:0" json:"actual"`
	UpdatedAt time.Time `json:"updated_at"`
}
