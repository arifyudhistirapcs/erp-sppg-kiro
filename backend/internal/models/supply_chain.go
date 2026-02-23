package models

import (
	"time"
)

// Supplier represents a vendor that supplies ingredients
type Supplier struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:200;not null;index" json:"name" validate:"required"`
	ContactPerson   string    `gorm:"size:100" json:"contact_person"`
	PhoneNumber     string    `gorm:"size:20" json:"phone_number"`
	Email           string    `gorm:"size:100" json:"email" validate:"omitempty,email"`
	Address         string    `gorm:"type:text" json:"address"`
	ProductCategory string    `gorm:"size:100;index" json:"product_category"`
	IsActive        bool      `gorm:"default:true;index" json:"is_active"`
	OnTimeDelivery  float64   `gorm:"default:0" json:"on_time_delivery"` // percentage
	QualityRating   float64   `gorm:"default:0" json:"quality_rating"`   // 1-5 scale
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PurchaseOrder represents an order placed with a supplier
type PurchaseOrder struct {
	ID               uint                `gorm:"primaryKey" json:"id"`
	PONumber         string              `gorm:"uniqueIndex;size:50;not null" json:"po_number" validate:"required"`
	SupplierID       uint                `gorm:"index;not null" json:"supplier_id"`
	OrderDate        time.Time           `gorm:"index;not null" json:"order_date"`
	ExpectedDelivery time.Time           `gorm:"index" json:"expected_delivery"`
	Status           string              `gorm:"size:20;not null;index" json:"status" validate:"required,oneof=pending approved received cancelled"` // pending, approved, received, cancelled
	TotalAmount      float64             `gorm:"not null" json:"total_amount"`
	ApprovedBy       *uint               `gorm:"index" json:"approved_by"`
	ApprovedAt       *time.Time          `json:"approved_at"`
	CreatedBy        uint                `gorm:"not null;index" json:"created_by"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Supplier         Supplier            `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Approver         *User               `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	Creator          User                `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	POItems          []PurchaseOrderItem `gorm:"foreignKey:POID" json:"po_items,omitempty"`
}

// PurchaseOrderItem represents a line item in a purchase order
type PurchaseOrderItem struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	POID         uint       `gorm:"index;not null" json:"po_id"`
	IngredientID uint       `gorm:"index;not null" json:"ingredient_id"`
	Quantity     float64    `gorm:"not null" json:"quantity" validate:"required,gt=0"`
	UnitPrice    float64    `gorm:"not null" json:"unit_price" validate:"required,gte=0"`
	Subtotal     float64    `gorm:"not null" json:"subtotal"`
	PO           PurchaseOrder `gorm:"foreignKey:POID" json:"po,omitempty"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// GoodsReceipt represents the receipt of goods from a supplier
type GoodsReceipt struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	GRNNumber     string             `gorm:"uniqueIndex;size:50;not null" json:"grn_number" validate:"required"`
	POID          uint               `gorm:"index;not null" json:"po_id"`
	ReceiptDate   time.Time          `gorm:"index;not null" json:"receipt_date"`
	InvoicePhoto  string             `gorm:"size:500" json:"invoice_photo"` // cloud storage URL
	ReceivedBy    uint               `gorm:"not null;index" json:"received_by"`
	Notes         string             `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time          `json:"created_at"`
	PurchaseOrder PurchaseOrder      `gorm:"foreignKey:POID" json:"purchase_order,omitempty"`
	Receiver      User               `gorm:"foreignKey:ReceivedBy" json:"receiver,omitempty"`
	GRNItems      []GoodsReceiptItem `gorm:"foreignKey:GRNID" json:"grn_items,omitempty"`
}

// GoodsReceiptItem represents a line item in a goods receipt
type GoodsReceiptItem struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	GRNID            uint       `gorm:"index;not null" json:"grn_id"`
	IngredientID     uint       `gorm:"index;not null" json:"ingredient_id"`
	OrderedQuantity  float64    `gorm:"not null" json:"ordered_quantity"`
	ReceivedQuantity float64    `gorm:"not null" json:"received_quantity" validate:"required,gte=0"`
	ExpiryDate       *time.Time `gorm:"index" json:"expiry_date"`
	GRN              GoodsReceipt `gorm:"foreignKey:GRNID" json:"grn,omitempty"`
	Ingredient       Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// InventoryItem represents current stock levels for an ingredient
type InventoryItem struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	IngredientID uint       `gorm:"uniqueIndex;not null" json:"ingredient_id"`
	Quantity     float64    `gorm:"not null" json:"quantity"`
	MinThreshold float64    `gorm:"not null" json:"min_threshold"`
	LastUpdated  time.Time  `gorm:"index;not null" json:"last_updated"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// InventoryMovement tracks all inventory changes
type InventoryMovement struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	IngredientID uint       `gorm:"index;not null" json:"ingredient_id"`
	MovementType string     `gorm:"size:20;not null;index" json:"movement_type" validate:"required,oneof=in out adjustment"` // in, out, adjustment
	Quantity     float64    `gorm:"not null" json:"quantity"`
	Reference    string     `gorm:"size:100;index" json:"reference"` // GRN number, recipe ID, etc.
	MovementDate time.Time  `gorm:"index;not null" json:"movement_date"`
	CreatedBy    uint       `gorm:"not null;index" json:"created_by"`
	Notes        string     `gorm:"type:text" json:"notes"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	Creator      User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
