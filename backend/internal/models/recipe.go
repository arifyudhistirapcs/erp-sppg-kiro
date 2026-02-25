package models

import (
	"time"
)

// Ingredient represents a raw material used in recipes
type Ingredient struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Code            string    `gorm:"size:20;index;default:''" json:"code"` // Auto-generated: B-XXXX
	Name            string    `gorm:"size:100;not null;index" json:"name" validate:"required"`
	Unit            string    `gorm:"size:20;not null" json:"unit" validate:"required"` // kg, liter, pcs, etc.
	CaloriesPer100g float64   `gorm:"default:0" json:"calories_per_100g"`
	ProteinPer100g  float64   `gorm:"default:0" json:"protein_per_100g"`
	CarbsPer100g    float64   `gorm:"default:0" json:"carbs_per_100g"`
	FatPer100g      float64   `gorm:"default:0" json:"fat_per_100g"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Recipe represents a food menu/recipe that consists of semi-finished goods
// Example: "Paket Ayam Goreng" consists of Nasi, Ayam Goreng, and Sambal
type Recipe struct {
	ID                   uint                    `gorm:"primaryKey" json:"id"`
	Name                 string                  `gorm:"size:200;not null;index" json:"name" validate:"required"`
	Category             string                  `gorm:"size:50;index" json:"category"`
	ServingSize          int                     `gorm:"not null" json:"serving_size" validate:"required,gt=0"` // number of portions
	Instructions         string                  `gorm:"type:text" json:"instructions"`
	TotalCalories        float64                 `gorm:"not null" json:"total_calories"`
	TotalProtein         float64                 `gorm:"not null" json:"total_protein"`
	TotalCarbs           float64                 `gorm:"not null" json:"total_carbs"`
	TotalFat             float64                 `gorm:"not null" json:"total_fat"`
	Version              int                     `gorm:"default:1;not null" json:"version"`
	IsActive             bool                    `gorm:"default:true;index" json:"is_active"`
	CreatedBy            uint                    `gorm:"not null;index" json:"created_by"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
	Creator              User                    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	
	// Relationships - Menu consists of SemiFinishedGoods
	RecipeItems []RecipeItem `gorm:"foreignKey:RecipeID" json:"recipe_items,omitempty"`
}

// RecipeItem represents semi-finished goods used in a menu/recipe
// Example: "Paket Ayam Goreng" uses 1 portion of Nasi, 1 portion of Ayam Goreng
type RecipeItem struct {
	ID                  uint              `gorm:"primaryKey" json:"id"`
	RecipeID            uint              `gorm:"index;not null" json:"recipe_id"`
	SemiFinishedGoodsID uint              `gorm:"index;not null" json:"semi_finished_goods_id"`
	Quantity            float64           `gorm:"not null" json:"quantity" validate:"required,gt=0"` // quantity of semi-finished goods
	
	// Relationships
	Recipe              Recipe            `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
	SemiFinishedGoods   SemiFinishedGoods `gorm:"foreignKey:SemiFinishedGoodsID" json:"semi_finished_goods,omitempty"`
}

// MenuPlan represents a weekly menu plan
type MenuPlan struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	WeekStart  time.Time  `gorm:"index;not null" json:"week_start"`
	WeekEnd    time.Time  `gorm:"not null" json:"week_end"`
	Status     string     `gorm:"size:20;not null;index" json:"status" validate:"required,oneof=draft approved"` // draft, approved
	ApprovedBy *uint      `gorm:"index" json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`
	CreatedBy  uint       `gorm:"not null;index" json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Approver   *User      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	Creator    User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	MenuItems  []MenuItem `gorm:"foreignKey:MenuPlanID" json:"menu_items,omitempty"`
}

// MenuItem represents a recipe assigned to a specific day in a menu plan
type MenuItem struct {
	ID                uint                        `gorm:"primaryKey" json:"id"`
	MenuPlanID        uint                        `gorm:"index;not null" json:"menu_plan_id"`
	Date              time.Time                   `gorm:"index;not null" json:"date"`
	RecipeID          uint                        `gorm:"index;not null" json:"recipe_id"`
	Portions          int                         `gorm:"not null" json:"portions" validate:"required,gt=0"`
	MenuPlan          MenuPlan                    `gorm:"foreignKey:MenuPlanID" json:"menu_plan,omitempty"`
	Recipe            Recipe                      `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
	SchoolAllocations []MenuItemSchoolAllocation  `gorm:"foreignKey:MenuItemID" json:"school_allocations,omitempty"`
}

// MenuItemSchoolAllocation represents portions of a menu item allocated to a specific school
type MenuItemSchoolAllocation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	MenuItemID uint      `gorm:"index;not null" json:"menu_item_id"`
	SchoolID   uint      `gorm:"index;not null" json:"school_id"`
	Portions   int       `gorm:"not null;check:portions > 0" json:"portions" validate:"required,gt=0"`
	Date       time.Time `gorm:"index;not null" json:"date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	// Relationships
	MenuItem   MenuItem  `gorm:"foreignKey:MenuItemID;constraint:OnDelete:CASCADE" json:"menu_item,omitempty"`
	School     School    `gorm:"foreignKey:SchoolID;constraint:OnDelete:RESTRICT" json:"school,omitempty"`
}

// RecipeVersion stores historical versions of recipes
// This is a snapshot of the recipe at a specific version
// Used for audit trail and version history tracking
type RecipeVersion struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	RecipeID             uint      `gorm:"index;not null" json:"recipe_id"`
	Version              int       `gorm:"not null;index" json:"version"`
	Name                 string    `gorm:"size:200;not null" json:"name"`
	Category             string    `gorm:"size:50" json:"category"`
	ServingSize          int       `gorm:"not null" json:"serving_size"`
	Instructions         string    `gorm:"type:text" json:"instructions"`
	TotalCalories        float64   `gorm:"not null" json:"total_calories"`
	TotalProtein         float64   `gorm:"not null" json:"total_protein"`
	TotalCarbs           float64   `gorm:"not null" json:"total_carbs"`
	TotalFat             float64   `gorm:"not null" json:"total_fat"`
	Changes              string    `gorm:"type:text" json:"changes"` // JSON array of changes
	CreatedBy            uint      `gorm:"not null" json:"created_by"`
	CreatedAt            time.Time `json:"created_at"`
	
	// Relationships
	Creator              User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
