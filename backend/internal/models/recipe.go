package models

import (
	"time"
)

// Ingredient represents a raw material used in recipes
type Ingredient struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:100;not null;index" json:"name" validate:"required"`
	Unit            string    `gorm:"size:20;not null" json:"unit" validate:"required"` // kg, liter, pcs, etc.
	CaloriesPer100g float64   `gorm:"not null" json:"calories_per_100g"`
	ProteinPer100g  float64   `gorm:"not null" json:"protein_per_100g"`
	CarbsPer100g    float64   `gorm:"not null" json:"carbs_per_100g"`
	FatPer100g      float64   `gorm:"not null" json:"fat_per_100g"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Recipe represents a food recipe with nutritional information
type Recipe struct {
	ID                uint               `gorm:"primaryKey" json:"id"`
	Name              string             `gorm:"size:200;not null;index" json:"name" validate:"required"`
	Category          string             `gorm:"size:50;index" json:"category"`
	ServingSize       int                `gorm:"not null" json:"serving_size" validate:"required,gt=0"` // number of portions
	Instructions      string             `gorm:"type:text" json:"instructions"`
	TotalCalories     float64            `gorm:"not null" json:"total_calories"`
	TotalProtein      float64            `gorm:"not null" json:"total_protein"`
	TotalCarbs        float64            `gorm:"not null" json:"total_carbs"`
	TotalFat          float64            `gorm:"not null" json:"total_fat"`
	Version           int                `gorm:"default:1;not null" json:"version"`
	IsActive          bool               `gorm:"default:true;index" json:"is_active"`
	CreatedBy         uint               `gorm:"not null;index" json:"created_by"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	Creator           User               `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	RecipeIngredients []RecipeIngredient `gorm:"foreignKey:RecipeID" json:"recipe_ingredients,omitempty"`
}

// RecipeIngredient represents the many-to-many relationship between recipes and ingredients
type RecipeIngredient struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	RecipeID     uint       `gorm:"index;not null" json:"recipe_id"`
	IngredientID uint       `gorm:"index;not null" json:"ingredient_id"`
	Quantity     float64    `gorm:"not null" json:"quantity" validate:"required,gt=0"` // in ingredient's unit
	Recipe       Recipe     `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
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
	ID         uint      `gorm:"primaryKey" json:"id"`
	MenuPlanID uint      `gorm:"index;not null" json:"menu_plan_id"`
	Date       time.Time `gorm:"index;not null" json:"date"`
	RecipeID   uint      `gorm:"index;not null" json:"recipe_id"`
	Portions   int       `gorm:"not null" json:"portions" validate:"required,gt=0"`
	MenuPlan   MenuPlan  `gorm:"foreignKey:MenuPlanID" json:"menu_plan,omitempty"`
	Recipe     Recipe    `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
}
