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
		&SemiFinishedMovement{},
		&Recipe{},
		&RecipeItem{},
		&RecipeVersion{},
		&MenuPlan{},
		&MenuItem{},
		&MenuItemSchoolAllocation{},
		
		// Supply Chain & Inventory
		&Supplier{},
		&PurchaseOrder{},
		&PurchaseOrderItem{},
		&GoodsReceipt{},
		&GoodsReceiptItem{},
		&InventoryItem{},
		&InventoryMovement{},
		&StokOpnameForm{},
		&StokOpnameItem{},
		
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
