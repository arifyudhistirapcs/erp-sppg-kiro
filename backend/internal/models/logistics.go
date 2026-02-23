package models

import (
	"time"
)

// School represents a school that receives food deliveries
type School struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:200;not null;index" json:"name" validate:"required"`
	Address       string    `gorm:"type:text" json:"address"`
	Latitude      float64   `gorm:"not null" json:"latitude" validate:"required,min=-90,max=90"`
	Longitude     float64   `gorm:"not null" json:"longitude" validate:"required,min=-180,max=180"`
	ContactPerson string    `gorm:"size:100" json:"contact_person"`
	PhoneNumber   string    `gorm:"size:20" json:"phone_number"`
	StudentCount  int       `gorm:"not null" json:"student_count" validate:"required,gte=0"`
	IsActive      bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DeliveryTask represents a delivery assignment for a driver
type DeliveryTask struct {
	ID         uint                 `gorm:"primaryKey" json:"id"`
	TaskDate   time.Time            `gorm:"index;not null" json:"task_date"`
	DriverID   uint                 `gorm:"index;not null" json:"driver_id"`
	SchoolID   uint                 `gorm:"index;not null" json:"school_id"`
	Portions   int                  `gorm:"not null" json:"portions" validate:"required,gt=0"`
	Status     string               `gorm:"size:20;not null;index" json:"status" validate:"required,oneof=pending in_progress completed cancelled"` // pending, in_progress, completed, cancelled
	RouteOrder int                  `gorm:"not null" json:"route_order"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
	Driver     User                 `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	School     School               `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	MenuItems  []DeliveryMenuItem   `gorm:"foreignKey:DeliveryTaskID" json:"menu_items,omitempty"`
}

// DeliveryMenuItem represents menu items included in a delivery
type DeliveryMenuItem struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	DeliveryTaskID uint         `gorm:"index;not null" json:"delivery_task_id"`
	RecipeID       uint         `gorm:"index;not null" json:"recipe_id"`
	Portions       int          `gorm:"not null" json:"portions" validate:"required,gt=0"`
	DeliveryTask   DeliveryTask `gorm:"foreignKey:DeliveryTaskID" json:"delivery_task,omitempty"`
	Recipe         Recipe       `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
}

// ElectronicPOD represents electronic proof of delivery
type ElectronicPOD struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	DeliveryTaskID uint         `gorm:"uniqueIndex;not null" json:"delivery_task_id"`
	PhotoURL       string       `gorm:"size:500" json:"photo_url"`
	SignatureURL   string       `gorm:"size:500" json:"signature_url"`
	Latitude       float64      `gorm:"not null" json:"latitude"`
	Longitude      float64      `gorm:"not null" json:"longitude"`
	RecipientName  string       `gorm:"size:100" json:"recipient_name"`
	OmprengDropOff int          `gorm:"not null" json:"ompreng_drop_off" validate:"gte=0"`
	OmprengPickUp  int          `gorm:"not null" json:"ompreng_pick_up" validate:"gte=0"`
	CompletedAt    time.Time    `gorm:"index;not null" json:"completed_at"`
	DeliveryTask   DeliveryTask `gorm:"foreignKey:DeliveryTaskID" json:"delivery_task,omitempty"`
}

// OmprengTracking tracks ompreng (food container) circulation
type OmprengTracking struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SchoolID   uint      `gorm:"index;not null" json:"school_id"`
	Date       time.Time `gorm:"index;not null" json:"date"`
	DropOff    int       `gorm:"not null" json:"drop_off" validate:"gte=0"`
	PickUp     int       `gorm:"not null" json:"pick_up" validate:"gte=0"`
	Balance    int       `gorm:"not null" json:"balance"` // cumulative balance at school
	RecordedBy uint      `gorm:"not null;index" json:"recorded_by"`
	CreatedAt  time.Time `json:"created_at"`
	School     School    `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Recorder   User      `gorm:"foreignKey:RecordedBy" json:"recorder,omitempty"`
}

// OmprengInventory tracks global ompreng inventory
type OmprengInventory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TotalOwned    int       `gorm:"not null" json:"total_owned" validate:"gte=0"`    // total ompreng owned by SPPG
	AtKitchen     int       `gorm:"not null" json:"at_kitchen" validate:"gte=0"`     // currently at central kitchen
	InCirculation int       `gorm:"not null" json:"in_circulation" validate:"gte=0"` // currently at schools
	Missing       int       `gorm:"not null" json:"missing" validate:"gte=0"`        // unaccounted for
	LastUpdated   time.Time `gorm:"index;not null" json:"last_updated"`
}
