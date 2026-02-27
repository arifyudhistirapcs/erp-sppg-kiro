package models

import (
	"time"
)

// School represents a school that receives food deliveries
type School struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	Name                 string    `gorm:"size:200;not null;index" json:"name" validate:"required"`
	Address              string    `gorm:"type:text" json:"address"`
	Latitude             float64   `gorm:"not null" json:"latitude" validate:"required,min=-90,max=90"`
	Longitude            float64   `gorm:"not null" json:"longitude" validate:"required,min=-180,max=180"`
	ContactPerson        string    `gorm:"size:100" json:"contact_person"`
	PhoneNumber          string    `gorm:"size:20" json:"phone_number"`
	StudentCount         int       `gorm:"not null" json:"student_count" validate:"required,gte=0"` // For SMP/SMA
	Category             string    `gorm:"size:10" json:"category" validate:"required,oneof=SD SMP SMA"`
	StudentCountGrade13  int       `gorm:"default:0" json:"student_count_grade_1_3"` // For SD only
	StudentCountGrade46  int       `gorm:"default:0" json:"student_count_grade_4_6"` // For SD only
	StaffCount           int       `gorm:"default:0" json:"staff_count" validate:"gte=0"`
	NPSN                 string    `gorm:"size:50" json:"npsn"`
	PrincipalName        string    `gorm:"size:255" json:"principal_name"`
	SchoolEmail          string    `gorm:"size:255" json:"school_email"`
	SchoolPhone          string    `gorm:"size:50" json:"school_phone"`
	CommitteeCount       int       `gorm:"default:0" json:"committee_count" validate:"gte=0"`
	CooperationLetterURL string    `gorm:"size:500" json:"cooperation_letter_url"`
	IsActive             bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// DeliveryTask represents a delivery assignment for a driver
type DeliveryTask struct {
	ID         uint               `gorm:"primaryKey" json:"id"`
	TaskDate   time.Time          `gorm:"index;not null" json:"task_date"`
	DriverID   uint               `gorm:"index;not null" json:"driver_id"`
	SchoolID   uint               `gorm:"index;not null" json:"school_id"`
	Portions   int                `gorm:"not null" json:"portions" validate:"required,gt=0"`
	Status     string             `gorm:"size:20;not null;index" json:"status" validate:"required,oneof=pending in_progress completed cancelled"` // pending, in_progress, completed, cancelled
	RouteOrder int                `gorm:"not null" json:"route_order"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	Driver     User               `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	School     School             `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	MenuItems  []DeliveryMenuItem `gorm:"foreignKey:DeliveryTaskID" json:"menu_items,omitempty"`
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

// DeliveryRecord represents a delivery record tracking menu delivery and ompreng lifecycle
type DeliveryRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DeliveryDate  time.Time `gorm:"index;not null" json:"delivery_date"`
	SchoolID      uint      `gorm:"index;not null" json:"school_id"`
	DriverID      uint      `gorm:"index;not null" json:"driver_id"`
	MenuItemID    uint      `gorm:"index;not null" json:"menu_item_id"`
	Portions      int       `gorm:"not null" json:"portions"`
	CurrentStatus string    `gorm:"size:50;not null;index" json:"current_status"`
	CurrentStage  int       `gorm:"not null;default:1;index" json:"current_stage"` // Stage number 1-16
	OmprengCount  int       `gorm:"not null" json:"ompreng_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	School        School    `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	Driver        User      `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	MenuItem      MenuItem  `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}

// StatusTransition represents a status change in the delivery lifecycle
type StatusTransition struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	DeliveryRecordID uint           `gorm:"index;not null" json:"delivery_record_id"`
	FromStatus       string         `gorm:"size:50" json:"from_status"`
	ToStatus         string         `gorm:"size:50;not null" json:"to_status"`
	Stage            int            `gorm:"not null;index" json:"stage"`        // Stage number 1-16
	TransitionedAt   time.Time      `gorm:"index;not null" json:"transitioned_at"`
	TransitionedBy   uint           `gorm:"index;not null" json:"transitioned_by"`
	Notes            string         `gorm:"type:text" json:"notes"`
	MediaURL         string         `gorm:"size:500" json:"media_url"`  // Photo/video URL
	MediaType        string         `gorm:"size:20" json:"media_type"`  // "photo" or "video"
	DeliveryRecord   DeliveryRecord `gorm:"foreignKey:DeliveryRecordID" json:"delivery_record,omitempty"`
	User             User           `gorm:"foreignKey:TransitionedBy" json:"user,omitempty"`
}

// OmprengCleaning represents ompreng cleaning tracking
type OmprengCleaning struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	DeliveryRecordID uint           `gorm:"index;not null" json:"delivery_record_id"`
	OmprengCount     int            `gorm:"not null" json:"ompreng_count"`
	CleaningStatus   string         `gorm:"size:30;not null" json:"cleaning_status"`
	StartedAt        *time.Time     `json:"started_at"`
	CompletedAt      *time.Time     `json:"completed_at"`
	CleanedBy        *uint          `gorm:"index" json:"cleaned_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeliveryRecord   DeliveryRecord `gorm:"foreignKey:DeliveryRecordID" json:"delivery_record,omitempty"`
	Cleaner          User           `gorm:"foreignKey:CleanedBy" json:"cleaner,omitempty"`
}

// DailySummary represents summary statistics for deliveries on a specific date
type DailySummary struct {
	TotalDeliveries     int            `json:"total_deliveries"`
	CompletedDeliveries int            `json:"completed_deliveries"`
	StatusCounts        map[string]int `json:"status_counts"`
	OmprengInCleaning   int            `json:"ompreng_in_cleaning"`
	OmprengCleaned      int            `json:"ompreng_cleaned"`
}
