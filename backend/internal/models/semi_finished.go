package models

import (
	"time"
)

// SemiFinishedGoods represents semi-finished products (barang setengah jadi)
// Examples: Nasi, Ayam Goreng, Sambal
type SemiFinishedGoods struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	Name                    string    `gorm:"size:100;not null;index" json:"name" validate:"required"`
	Unit                    string    `gorm:"size:20;not null" json:"unit" validate:"required"` // kg, liter, pcs, etc.
	Category                string    `gorm:"size:50;index" json:"category"` // nasi, lauk, sambal, etc.
	Description             string    `gorm:"type:text" json:"description"`
	CaloriesPer100g         float64   `gorm:"not null" json:"calories_per_100g"`
	ProteinPer100g          float64   `gorm:"not null" json:"protein_per_100g"`
	CarbsPer100g            float64   `gorm:"not null" json:"carbs_per_100g"`
	FatPer100g              float64   `gorm:"not null" json:"fat_per_100g"`
	QuantityPerPortionSmall float64   `gorm:"default:0" json:"quantity_per_portion_small"` // gram needed for 1 small portion (e.g., 50g nasi)
	QuantityPerPortionLarge float64   `gorm:"default:0" json:"quantity_per_portion_large"` // gram needed for 1 large portion (e.g., 100g nasi)
	StockQuantity           float64   `gorm:"default:0" json:"stock_quantity"`
	MinThreshold            float64   `gorm:"default:10" json:"min_threshold"`
	IsActive                bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	
	// Relationships
	Recipe            *SemiFinishedRecipe           `gorm:"foreignKey:SemiFinishedGoodsID" json:"recipe,omitempty"`
}

// SemiFinishedRecipe represents the recipe to produce semi-finished goods from raw ingredients
// Example: Recipe to make "Nasi" from "Beras"
type SemiFinishedRecipe struct {
	ID                   uint                          `gorm:"primaryKey" json:"id"`
	SemiFinishedGoodsID  uint                          `gorm:"uniqueIndex;not null" json:"semi_finished_goods_id"`
	Name                 string                        `gorm:"size:200;not null" json:"name"`
	Instructions         string                        `gorm:"type:text" json:"instructions"`
	YieldAmount          float64                       `gorm:"not null" json:"yield_amount"` // How much semi-finished product is produced (e.g., 1 kg Nasi from 500g Beras)
	IsActive             bool                          `gorm:"default:true" json:"is_active"`
	CreatedBy            uint                          `gorm:"not null" json:"created_by"`
	CreatedAt            time.Time                     `json:"created_at"`
	UpdatedAt            time.Time                     `json:"updated_at"`
	
	// Relationships
	SemiFinishedGoods  SemiFinishedGoods               `gorm:"foreignKey:SemiFinishedGoodsID" json:"semi_finished_goods,omitempty"`
	Ingredients        []SemiFinishedRecipeIngredient  `gorm:"foreignKey:SemiFinishedRecipeID" json:"ingredients,omitempty"`
	Creator            User                            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// SemiFinishedRecipeIngredient represents ingredients needed to produce semi-finished goods
// Example: To make Nasi, need 500g Beras
type SemiFinishedRecipeIngredient struct {
	ID                    uint              `gorm:"primaryKey" json:"id"`
	SemiFinishedRecipeID  uint              `gorm:"index;not null" json:"semi_finished_recipe_id"`
	IngredientID          uint              `gorm:"index;not null" json:"ingredient_id"`
	Quantity              float64           `gorm:"not null" json:"quantity" validate:"required,gt=0"`
	
	// Relationships
	SemiFinishedRecipe  SemiFinishedRecipe  `gorm:"foreignKey:SemiFinishedRecipeID" json:"semi_finished_recipe,omitempty"`
	Ingredient          Ingredient          `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// SemiFinishedInventory tracks stock of semi-finished goods
type SemiFinishedInventory struct {
	ID                  uint              `gorm:"primaryKey" json:"id"`
	SemiFinishedGoodsID uint              `gorm:"uniqueIndex;not null" json:"semi_finished_goods_id"`
	Quantity            float64           `gorm:"not null" json:"quantity"`
	MinThreshold        float64           `gorm:"not null" json:"min_threshold"`
	LastUpdated         time.Time         `gorm:"index;not null" json:"last_updated"`
	
	// Relationships
	SemiFinishedGoods   SemiFinishedGoods `gorm:"foreignKey:SemiFinishedGoodsID" json:"semi_finished_goods,omitempty"`
}

// SemiFinishedProductionLog tracks production of semi-finished goods
type SemiFinishedProductionLog struct {
	ID                  uint              `gorm:"primaryKey" json:"id"`
	SemiFinishedGoodsID uint              `gorm:"index;not null" json:"semi_finished_goods_id"`
	Quantity            float64           `gorm:"not null" json:"quantity"`
	ProductionDate      time.Time         `gorm:"index;not null" json:"production_date"`
	CreatedBy           uint              `gorm:"not null" json:"created_by"`
	Notes               string            `gorm:"type:text" json:"notes"`
	CreatedAt           time.Time         `json:"created_at"`
	
	// Relationships
	SemiFinishedGoods   SemiFinishedGoods `gorm:"foreignKey:SemiFinishedGoodsID" json:"semi_finished_goods,omitempty"`
	Creator             User              `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
