package models

// AllModels returns a slice of all model types for migration
func AllModels() []interface{} {
	return []interface{}{
		// User & Authentication
		&User{},
		&AuditTrail{},
		
		// Recipe & Menu Planning - Ingredients & Semi-Finished Goods
		&Ingredient{},
		&SemiFinishedGoods{},
		&SemiFinishedRecipe{},
		&SemiFinishedRecipeIngredient{},
		&SemiFinishedInventory{},
		&SemiFinishedProductionLog{},
		&Recipe{},
		&RecipeItem{},
		&RecipeVersion{},
		&MenuPlan{},
		&MenuItem{},
		
		// Supply Chain & Inventory
		&Supplier{},
		&PurchaseOrder{},
		&PurchaseOrderItem{},
		&GoodsReceipt{},
		&GoodsReceiptItem{},
		&InventoryItem{},
		&InventoryMovement{},
		
		// Logistics & Distribution
		&School{},
		&DeliveryTask{},
		&DeliveryMenuItem{},
		&ElectronicPOD{},
		&OmprengTracking{},
		&OmprengInventory{},
		
		// Human Resources
		&Employee{},
		&Attendance{},
		&WiFiConfig{},
		
		// Financial & Asset Management
		&KitchenAsset{},
		&AssetMaintenance{},
		&CashFlowEntry{},
		&BudgetTarget{},
		
		// System Configuration
		&SystemConfig{},
		&Notification{},
	}
}
