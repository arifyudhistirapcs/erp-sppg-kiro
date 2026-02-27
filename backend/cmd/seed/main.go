package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/models"
	"github.com/erp-sppg/backend/internal/utils"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Starting database seeding...")

	// Seed all data
	seedUsers(db)
	seedIngredients(db)
	seedSemiFinishedGoods(db)
	seedRecipes(db)
	seedSuppliers(db)
	seedPurchaseOrders(db)
	seedInventory(db)
	seedSchools(db)
	seedEmployees(db)
	seedKitchenAssets(db)
	seedCashFlowEntries(db)
	seedBudgetTargets(db)
	seedSystemConfig(db)
	seedMenuPlans(db)
	seedDeliveryTasks(db)
	seedOmprengTracking(db)
	seedNotifications(db)
	seedAuditTrails(db)
	seedWiFiConfig(db)

	log.Println("Database seeding completed successfully!")
}

func seedUsers(db *gorm.DB) {
	log.Println("Seeding users...")

	users := []models.User{
		{NIK: "12345678901", Email: "kepala.sppg@sppg.com", FullName: "Kepala SPPG", PhoneNumber: "081234567890", Role: "kepala_sppg", IsActive: true},
		{NIK: "12345678902", Email: "kepala.yayasan@sppg.com", FullName: "Kepala Yayasan", PhoneNumber: "081234567891", Role: "kepala_yayasan", IsActive: true},
		{NIK: "12345678903", Email: "akuntan@sppg.com", FullName: "Akuntan", PhoneNumber: "081234567892", Role: "akuntan", IsActive: true},
		{NIK: "12345678904", Email: "ahli.gizi@sppg.com", FullName: "Ahli Gizi", PhoneNumber: "081234567893", Role: "ahli_gizi", IsActive: true},
		{NIK: "12345678905", Email: "pengadaan@sppg.com", FullName: "Staff Pengadaan", PhoneNumber: "081234567894", Role: "pengadaan", IsActive: true},
		{NIK: "12345678906", Email: "chef@sppg.com", FullName: "Chef Utama", PhoneNumber: "081234567895", Role: "chef", IsActive: true},
		{NIK: "12345678907", Email: "packing@sppg.com", FullName: "Staff Packing", PhoneNumber: "081234567896", Role: "packing", IsActive: true},
		{NIK: "12345678908", Email: "driver1@sppg.com", FullName: "Driver 1", PhoneNumber: "081234567897", Role: "driver", IsActive: true},
		{NIK: "12345678909", Email: "driver2@sppg.com", FullName: "Driver 2", PhoneNumber: "081234567898", Role: "driver", IsActive: true},
		{NIK: "12345678910", Email: "asisten@sppg.com", FullName: "Asisten Lapangan", PhoneNumber: "081234567899", Role: "asisten_lapangan", IsActive: true},
	}

	for i := range users {
		hash, _ := utils.HashPassword("password123")
		users[i].PasswordHash = hash
		db.FirstOrCreate(&users[i], models.User{Email: users[i].Email})
	}

	log.Printf("Seeded %d users\n", len(users))
}

func seedIngredients(db *gorm.DB) {
	log.Println("Seeding ingredients...")

	ingredients := []models.Ingredient{
		{Name: "Beras Putih", Unit: "kg", CaloriesPer100g: 365, ProteinPer100g: 7.5, CarbsPer100g: 80, FatPer100g: 0.5},
		{Name: "Daging Ayam", Unit: "kg", CaloriesPer100g: 165, ProteinPer100g: 31, CarbsPer100g: 0, FatPer100g: 3.6},
		{Name: "Telur", Unit: "kg", CaloriesPer100g: 155, ProteinPer100g: 13, CarbsPer100g: 1.1, FatPer100g: 11},
		{Name: "Minyak Goreng", Unit: "liter", CaloriesPer100g: 884, ProteinPer100g: 0, CarbsPer100g: 0, FatPer100g: 100},
		{Name: "Garam", Unit: "kg", CaloriesPer100g: 0, ProteinPer100g: 0, CarbsPer100g: 0, FatPer100g: 0},
		{Name: "Gula Pasir", Unit: "kg", CaloriesPer100g: 387, ProteinPer100g: 0, CarbsPer100g: 100, FatPer100g: 0},
		{Name: "Cabai Merah", Unit: "kg", CaloriesPer100g: 40, ProteinPer100g: 1.9, CarbsPer100g: 8.8, FatPer100g: 0.4},
		{Name: "Bawang Merah", Unit: "kg", CaloriesPer100g: 40, ProteinPer100g: 1.1, CarbsPer100g: 9.3, FatPer100g: 0.1},
		{Name: "Bawang Putih", Unit: "kg", CaloriesPer100g: 149, ProteinPer100g: 6.4, CarbsPer100g: 33, FatPer100g: 0.5},
		{Name: "Santan", Unit: "liter", CaloriesPer100g: 230, ProteinPer100g: 2.3, CarbsPer100g: 5.5, FatPer100g: 24},
		{Name: "Tahu", Unit: "kg", CaloriesPer100g: 76, ProteinPer100g: 8.1, CarbsPer100g: 1.9, FatPer100g: 4.8},
		{Name: "Tempe", Unit: "kg", CaloriesPer100g: 193, ProteinPer100g: 19, CarbsPer100g: 9.4, FatPer100g: 11},
		{Name: "Wortel", Unit: "kg", CaloriesPer100g: 41, ProteinPer100g: 0.9, CarbsPer100g: 9.6, FatPer100g: 0.2},
		{Name: "Kacang Panjang", Unit: "kg", CaloriesPer100g: 47, ProteinPer100g: 2.6, CarbsPer100g: 8.9, FatPer100g: 0.1},
		{Name: "Kangkung", Unit: "kg", CaloriesPer100g: 19, ProteinPer100g: 2.6, CarbsPer100g: 3.1, FatPer100g: 0.2},
		{Name: "Daging Sapi", Unit: "kg", CaloriesPer100g: 250, ProteinPer100g: 26, CarbsPer100g: 0, FatPer100g: 15},
		{Name: "Ikan Lele", Unit: "kg", CaloriesPer100g: 105, ProteinPer100g: 18, CarbsPer100g: 0, FatPer100g: 3.5},
		{Name: "Tepung Terigu", Unit: "kg", CaloriesPer100g: 364, ProteinPer100g: 10, CarbsPer100g: 76, FatPer100g: 1},
	}

	for i := range ingredients {
		db.FirstOrCreate(&ingredients[i], models.Ingredient{Name: ingredients[i].Name})
	}

	log.Printf("Seeded %d ingredients\n", len(ingredients))
}

func seedSemiFinishedGoods(db *gorm.DB) {
	log.Println("Seeding semi-finished goods...")

	goods := []models.SemiFinishedGoods{
		{Name: "Nasi Putih", Unit: "kg", Category: "nasi", Description: "Nasi putih matang", CaloriesPer100g: 175, ProteinPer100g: 3.5, CarbsPer100g: 38, FatPer100g: 0.3, StockQuantity: 50, MinThreshold: 20, IsActive: true},
		{Name: "Nasi Uduk", Unit: "kg", Category: "nasi", Description: "Nasi uduk dengan santan", CaloriesPer100g: 200, ProteinPer100g: 3.8, CarbsPer100g: 35, FatPer100g: 4.5, StockQuantity: 30, MinThreshold: 15, IsActive: true},
		{Name: "Ayam Goreng", Unit: "kg", Category: "lauk", Description: "Ayam goreng crispy", CaloriesPer100g: 280, ProteinPer100g: 25, CarbsPer100g: 5, FatPer100g: 18, StockQuantity: 25, MinThreshold: 10, IsActive: true},
		{Name: "Ayam Bumbu Kuning", Unit: "kg", Category: "lauk", Description: "Ayam masak bumbu kuning", CaloriesPer100g: 220, ProteinPer100g: 28, CarbsPer100g: 3, FatPer100g: 10, StockQuantity: 20, MinThreshold: 10, IsActive: true},
		{Name: "Telur Dadar", Unit: "kg", Category: "lauk", Description: "Telur dadar goreng", CaloriesPer100g: 180, ProteinPer100g: 12, CarbsPer100g: 2, FatPer100g: 14, StockQuantity: 15, MinThreshold: 8, IsActive: true},
		{Name: "Tempe Goreng", Unit: "kg", Category: "lauk", Description: "Tempe goreng crispy", CaloriesPer100g: 250, ProteinPer100g: 18, CarbsPer100g: 12, FatPer100g: 16, StockQuantity: 20, MinThreshold: 10, IsActive: true},
		{Name: "Tahu Goreng", Unit: "kg", Category: "lauk", Description: "Tahu goreng renyah", CaloriesPer100g: 150, ProteinPer100g: 10, CarbsPer100g: 5, FatPer100g: 11, StockQuantity: 18, MinThreshold: 10, IsActive: true},
		{Name: "Rendang Sapi", Unit: "kg", Category: "lauk", Description: "Rendang daging sapi", CaloriesPer100g: 300, ProteinPer100g: 22, CarbsPer100g: 4, FatPer100g: 22, StockQuantity: 15, MinThreshold: 8, IsActive: true},
		{Name: "Sambal Merah", Unit: "kg", Category: "sambal", Description: "Sambal cabai merah", CaloriesPer100g: 120, ProteinPer100g: 2, CarbsPer100g: 15, FatPer100g: 8, StockQuantity: 10, MinThreshold: 5, IsActive: true},
		{Name: "Sambal Ijo", Unit: "kg", Category: "sambal", Description: "Sambal cabai hijau", CaloriesPer100g: 100, ProteinPer100g: 1.8, CarbsPer100g: 12, FatPer100g: 7, StockQuantity: 8, MinThreshold: 4, IsActive: true},
		{Name: "Tumis Kangkung", Unit: "kg", Category: "sayur", Description: "Kangkung tumis bawang putih", CaloriesPer100g: 80, ProteinPer100g: 4, CarbsPer100g: 8, FatPer100g: 4, StockQuantity: 15, MinThreshold: 8, IsActive: true},
		{Name: "Capcay", Unit: "kg", Category: "sayur", Description: "Capcay sayuran", CaloriesPer100g: 70, ProteinPer100g: 3.5, CarbsPer100g: 10, FatPer100g: 2, StockQuantity: 12, MinThreshold: 6, IsActive: true},
		{Name: "Sayur Asem", Unit: "liter", Category: "lauk_berkuah", Description: "Sayur asem segar", CaloriesPer100g: 45, ProteinPer100g: 1.5, CarbsPer100g: 8, FatPer100g: 1, StockQuantity: 20, MinThreshold: 10, IsActive: true},
		{Name: "Sop Ayam", Unit: "liter", Category: "lauk_berkuah", Description: "Sop ayam hangat", CaloriesPer100g: 65, ProteinPer100g: 5, CarbsPer100g: 6, FatPer100g: 2.5, StockQuantity: 18, MinThreshold: 10, IsActive: true},
		{Name: "Es Teh Manis", Unit: "liter", Category: "minuman", Description: "Es teh manis", CaloriesPer100g: 40, ProteinPer100g: 0, CarbsPer100g: 10, FatPer100g: 0, StockQuantity: 30, MinThreshold: 15, IsActive: true},
		{Name: "Air Mineral", Unit: "liter", Category: "minuman", Description: "Air mineral", CaloriesPer100g: 0, ProteinPer100g: 0, CarbsPer100g: 0, FatPer100g: 0, StockQuantity: 100, MinThreshold: 30, IsActive: true},
	}

	for i := range goods {
		db.FirstOrCreate(&goods[i], models.SemiFinishedGoods{Name: goods[i].Name})
	}

	log.Printf("Seeded %d semi-finished goods\n", len(goods))

	// Seed semi-finished recipes and inventory
	seedSemiFinishedRecipes(db, goods)
}

func seedSemiFinishedRecipes(db *gorm.DB, goods []models.SemiFinishedGoods) {
	log.Println("Seeding semi-finished recipes...")

	var ingredients []models.Ingredient
	db.Find(&ingredients)
	if len(ingredients) == 0 {
		log.Println("No ingredients found, skipping recipe creation")
		return
	}

	var kepalaSPPG models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)

	for _, g := range goods {
		// Create recipe
		recipe := models.SemiFinishedRecipe{
			SemiFinishedGoodsID: g.ID,
			Name:                "Resep " + g.Name,
			Instructions:        "Cara membuat " + g.Name + ": 1. Siapkan bahan 2. Proses sesuai SOP",
			YieldAmount:         10,
			IsActive:            true,
			CreatedBy:           kepalaSPPG.ID,
		}
		db.FirstOrCreate(&recipe, models.SemiFinishedRecipe{SemiFinishedGoodsID: g.ID})

		// Create recipe ingredients (2-4 random ingredients)
		numIngredients := rand.Intn(3) + 2
		for j := 0; j < numIngredients && j < len(ingredients); j++ {
			ing := ingredients[rand.Intn(len(ingredients))]
			recipeIng := models.SemiFinishedRecipeIngredient{
				SemiFinishedRecipeID: recipe.ID,
				IngredientID:         ing.ID,
				Quantity:             float64(rand.Intn(5)+1) * 0.5,
			}
			db.FirstOrCreate(&recipeIng, models.SemiFinishedRecipeIngredient{
				SemiFinishedRecipeID: recipe.ID,
				IngredientID:         ing.ID,
			})
		}

		// Create inventory record
		inv := models.SemiFinishedInventory{
			SemiFinishedGoodsID: g.ID,
			Quantity:            g.StockQuantity,
			MinThreshold:        g.MinThreshold,
			LastUpdated:         time.Now(),
		}
		db.FirstOrCreate(&inv, models.SemiFinishedInventory{SemiFinishedGoodsID: g.ID})
	}

	log.Printf("Seeded recipes and inventory for %d semi-finished goods\n", len(goods))
}

func seedRecipes(db *gorm.DB) {
	log.Println("Seeding recipes...")

	var semiFinishedGoods []models.SemiFinishedGoods
	db.Find(&semiFinishedGoods)

	var kepalaSPPG models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)

	recipes := []struct {
		name  string
		items []struct{ name string; qty float64 }
	}{
		{
			name: "Paket Ayam Goreng Komplit",
			items: []struct{ name string; qty float64 }{
				{"Nasi Putih", 0.2}, {"Ayam Goreng", 0.15}, {"Sambal Merah", 0.03}, {"Tumis Kangkung", 0.1}, {"Es Teh Manis", 0.25},
			},
		},
		{
			name: "Paket Rendang Spesial",
			items: []struct{ name string; qty float64 }{
				{"Nasi Putih", 0.2}, {"Rendang Sapi", 0.12}, {"Sambal Ijo", 0.03}, {"Capcay", 0.1}, {"Es Teh Manis", 0.25},
			},
		},
		{
			name: "Paket Ayam Bumbu Kuning",
			items: []struct{ name string; qty float64 }{
				{"Nasi Uduk", 0.2}, {"Ayam Bumbu Kuning", 0.15}, {"Sambal Merah", 0.03}, {"Tahu Goreng", 0.1}, {"Es Teh Manis", 0.25},
			},
		},
		{
			name: "Paket Lauk Vegetarian",
			items: []struct{ name string; qty float64 }{
				{"Nasi Putih", 0.2}, {"Tempe Goreng", 0.12}, {"Tahu Goreng", 0.1}, {"Sambal Merah", 0.03}, {"Tumis Kangkung", 0.1}, {"Es Teh Manis", 0.25},
			},
		},
		{
			name: "Paket Sop Ayam Hangat",
			items: []struct{ name string; qty float64 }{
				{"Nasi Putih", 0.15}, {"Sop Ayam", 0.3}, {"Telur Dadar", 0.1}, {"Sambal Ijo", 0.03}, {"Es Teh Manis", 0.25},
			},
		},
		{
			name: "Paket Lele Goreng",
			items: []struct{ name string; qty float64 }{
				{"Nasi Putih", 0.2}, {"Tempe Goreng", 0.1}, {"Sambal Merah", 0.03}, {"Sayur Asem", 0.25}, {"Es Teh Manis", 0.25},
			},
		},
	}

	for _, r := range recipes {
		recipe := models.Recipe{
			Name:         r.name,
			Category:     "paket",
			// ServingSize:  1, // Removed - nutrition is now per menu
			Instructions: "Siapkan semua komponen dan susun di piring sesuai porsi",
			IsActive:     true,
			CreatedBy:    kepalaSPPG.ID,
		}

		// Calculate nutrition
		var totalCal, totalProtein, totalCarbs, totalFat float64
		for _, item := range r.items {
			for _, sf := range semiFinishedGoods {
				if sf.Name == item.name {
					totalCal += sf.CaloriesPer100g * item.qty * 10
					totalProtein += sf.ProteinPer100g * item.qty * 10
					totalCarbs += sf.CarbsPer100g * item.qty * 10
					totalFat += sf.FatPer100g * item.qty * 10
					break
				}
			}
		}
		recipe.TotalCalories = totalCal
		recipe.TotalProtein = totalProtein
		recipe.TotalCarbs = totalCarbs
		recipe.TotalFat = totalFat

		db.FirstOrCreate(&recipe, models.Recipe{Name: recipe.Name})

		// Create recipe items
		for _, item := range r.items {
			for _, sf := range semiFinishedGoods {
				if sf.Name == item.name {
					recipeItem := models.RecipeItem{
						RecipeID:            recipe.ID,
						SemiFinishedGoodsID: sf.ID,
						Quantity:            item.qty,
					}
					db.FirstOrCreate(&recipeItem, models.RecipeItem{RecipeID: recipe.ID, SemiFinishedGoodsID: sf.ID})
					break
				}
			}
		}
	}

	log.Printf("Seeded %d recipes\n", len(recipes))
}

func seedSuppliers(db *gorm.DB) {
	log.Println("Seeding suppliers...")

	suppliers := []models.Supplier{
		{Name: "PT Beras Jaya", ContactPerson: "Budi Santoso", PhoneNumber: "08111111111", Email: "budi@berasjaya.com", Address: "Jl. Padi No. 1, Jakarta", ProductCategory: "Beras & Bahan Pokok", IsActive: true, OnTimeDelivery: 95, QualityRating: 4.5},
		{Name: "CV Ayam Segar", ContactPerson: "Siti Aminah", PhoneNumber: "08122222222", Email: "siti@ayamsegar.com", Address: "Jl. Unggas No. 5, Bogor", ProductCategory: "Daging & Unggas", IsActive: true, OnTimeDelivery: 90, QualityRating: 4.2},
		{Name: "UD Sayur Segar", ContactPerson: "Ahmad Yani", PhoneNumber: "08133333333", Email: "ahmad@sayursegar.com", Address: "Jl. Sayur No. 10, Depok", ProductCategory: "Sayuran", IsActive: true, OnTimeDelivery: 88, QualityRating: 4.0},
		{Name: "PT Minyak Goreng Nusantara", ContactPerson: "Dewi Kusuma", PhoneNumber: "08144444444", Email: "dewi@minyaknusantara.com", Address: "Jl. Industri No. 15, Tangerang", ProductCategory: "Minyak & Bumbu", IsActive: true, OnTimeDelivery: 98, QualityRating: 4.7},
		{Name: "CV Telur Berkah", ContactPerson: "Hendra Wijaya", PhoneNumber: "08155555555", Email: "hendra@telurberkah.com", Address: "Jl. Telur No. 20, Bekasi", ProductCategory: "Telur & Protein", IsActive: true, OnTimeDelivery: 92, QualityRating: 4.3},
		{Name: "PT Tahu Tempe Makmur", ContactPerson: "Ratna Sari", PhoneNumber: "08166666666", Email: "ratna@tahutempe.com", Address: "Jl. Kedelai No. 25, Jakarta", ProductCategory: "Protein Nabati", IsActive: true, OnTimeDelivery: 85, QualityRating: 3.9},
	}

	for i := range suppliers {
		db.FirstOrCreate(&suppliers[i], models.Supplier{Name: suppliers[i].Name})
	}

	log.Printf("Seeded %d suppliers\n", len(suppliers))
}

func seedPurchaseOrders(db *gorm.DB) {
	log.Println("Seeding purchase orders...")

	var suppliers []models.Supplier
	db.Find(&suppliers)

	var ingredients []models.Ingredient
	db.Find(&ingredients)

	var kepalaSPPG models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)

	statuses := []string{"pending", "approved", "received"}

	for i := 0; i < 10; i++ {
		supplier := suppliers[rand.Intn(len(suppliers))]
		status := statuses[rand.Intn(len(statuses))]

		po := models.PurchaseOrder{
			PONumber:         fmt.Sprintf("PO-%06d", i+1),
			SupplierID:       supplier.ID,
			OrderDate:        time.Now().AddDate(0, 0, -rand.Intn(30)),
			ExpectedDelivery: time.Now().AddDate(0, 0, rand.Intn(7)),
			Status:           status,
			TotalAmount:      float64(rand.Intn(5000000) + 1000000),
			CreatedBy:        kepalaSPPG.ID,
		}

		if status == "approved" || status == "received" {
			now := time.Now()
			po.ApprovedBy = &kepalaSPPG.ID
			po.ApprovedAt = &now
		}

		db.FirstOrCreate(&po, models.PurchaseOrder{PONumber: po.PONumber})

		// Create PO items
		numItems := rand.Intn(3) + 1
		for j := 0; j < numItems; j++ {
			ing := ingredients[rand.Intn(len(ingredients))]
			qty := float64(rand.Intn(50) + 10)
			price := float64(rand.Intn(50000) + 10000)

			poItem := models.PurchaseOrderItem{
				POID:         po.ID,
				IngredientID: ing.ID,
				Quantity:     qty,
				UnitPrice:    price,
				Subtotal:     qty * price,
			}
			db.FirstOrCreate(&poItem, models.PurchaseOrderItem{POID: po.ID, IngredientID: ing.ID})
		}
	}

	log.Println("Seeded 10 purchase orders")
}

func seedInventory(db *gorm.DB) {
	log.Println("Seeding inventory...")

	var ingredients []models.Ingredient
	db.Find(&ingredients)

	for _, ing := range ingredients {
		inv := models.InventoryItem{
			IngredientID: ing.ID,
			Quantity:     float64(rand.Intn(100) + 20),
			MinThreshold: float64(rand.Intn(20) + 5),
			LastUpdated:  time.Now(),
		}
		db.FirstOrCreate(&inv, models.InventoryItem{IngredientID: ing.ID})

		// Create inventory movement
		movement := models.InventoryMovement{
			IngredientID: ing.ID,
			MovementType: "in",
			Quantity:     inv.Quantity,
			Reference:    "Initial Stock",
			MovementDate: time.Now(),
			CreatedBy:    1,
			Notes:        "Initial inventory seed",
		}
		db.Create(&movement)
	}

	log.Printf("Seeded inventory for %d ingredients\n", len(ingredients))
}

func seedSchools(db *gorm.DB) {
	log.Println("Seeding schools...")

	schools := []models.School{
		{
			Name:                 "SD Negeri 1 Jakarta",
			Address:              "Jl. Pendidikan No. 1, Jakarta",
			Latitude:             -6.2088,
			Longitude:            106.8456,
			ContactPerson:        "Pak Sumarno",
			PhoneNumber:          "08211111111",
			StudentCount:         250,
			Category:             "SD",
			StudentCountGrade13:  130,
			StudentCountGrade46:  120,
			StaffCount:           25,
			NPSN:                 "20100001",
			PrincipalName:        "Drs. Sumarno, M.Pd",
			SchoolEmail:          "sdn1jakarta@gmail.com",
			SchoolPhone:          "021-1111111",
			CommitteeCount:       15,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SD Negeri 2 Bogor",
			Address:              "Jl. Pendidikan No. 2, Bogor",
			Latitude:             -6.5944,
			Longitude:            106.7892,
			ContactPerson:        "Ibu Sutarti",
			PhoneNumber:          "08222222222",
			StudentCount:         180,
			Category:             "SD",
			StudentCountGrade13:  95,
			StudentCountGrade46:  85,
			StaffCount:           20,
			NPSN:                 "20100002",
			PrincipalName:        "Sutarti, S.Pd",
			SchoolEmail:          "sdn2bogor@gmail.com",
			SchoolPhone:          "0251-2222222",
			CommitteeCount:       12,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SMP Negeri 1 Depok",
			Address:              "Jl. Pendidikan No. 3, Depok",
			Latitude:             -6.3856,
			Longitude:            106.8307,
			ContactPerson:        "Pak Joko",
			PhoneNumber:          "08233333333",
			StudentCount:         320,
			Category:             "SMP",
			StudentCountGrade13:  0,
			StudentCountGrade46:  0,
			StaffCount:           35,
			NPSN:                 "20200001",
			PrincipalName:        "Joko Widodo, M.Pd",
			SchoolEmail:          "smpn1depok@gmail.com",
			SchoolPhone:          "021-3333333",
			CommitteeCount:       18,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SD Negeri 4 Tangerang",
			Address:              "Jl. Pendidikan No. 4, Tangerang",
			Latitude:             -6.1702,
			Longitude:            106.6403,
			ContactPerson:        "Ibu Ratna",
			PhoneNumber:          "08244444444",
			StudentCount:         300,
			Category:             "SD",
			StudentCountGrade13:  155,
			StudentCountGrade46:  145,
			StaffCount:           28,
			NPSN:                 "20100004",
			PrincipalName:        "Ratna Dewi, S.Pd, M.Pd",
			SchoolEmail:          "sdn4tangerang@gmail.com",
			SchoolPhone:          "021-4444444",
			CommitteeCount:       16,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SMA Negeri 1 Bekasi",
			Address:              "Jl. Pendidikan No. 5, Bekasi",
			Latitude:             -6.2349,
			Longitude:            106.9896,
			ContactPerson:        "Pak Widodo",
			PhoneNumber:          "08255555555",
			StudentCount:         450,
			Category:             "SMA",
			StudentCountGrade13:  0,
			StudentCountGrade46:  0,
			StaffCount:           45,
			NPSN:                 "20300001",
			PrincipalName:        "Dr. Widodo Santoso, M.Pd",
			SchoolEmail:          "sman1bekasi@gmail.com",
			SchoolPhone:          "021-5555555",
			CommitteeCount:       20,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SD Negeri 6 Jakarta",
			Address:              "Jl. Pendidikan No. 6, Jakarta",
			Latitude:             -6.2500,
			Longitude:            106.8000,
			ContactPerson:        "Ibu Maryati",
			PhoneNumber:          "08266666666",
			StudentCount:         275,
			Category:             "SD",
			StudentCountGrade13:  140,
			StudentCountGrade46:  135,
			StaffCount:           26,
			NPSN:                 "20100006",
			PrincipalName:        "Maryati, S.Pd",
			SchoolEmail:          "sdn6jakarta@gmail.com",
			SchoolPhone:          "021-6666666",
			CommitteeCount:       14,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SMP Negeri 2 Jakarta",
			Address:              "Jl. Pendidikan No. 7, Jakarta",
			Latitude:             -6.2200,
			Longitude:            106.8300,
			ContactPerson:        "Pak Bambang",
			PhoneNumber:          "08277777777",
			StudentCount:         280,
			Category:             "SMP",
			StudentCountGrade13:  0,
			StudentCountGrade46:  0,
			StaffCount:           32,
			NPSN:                 "20200002",
			PrincipalName:        "Bambang Suryanto, M.Pd",
			SchoolEmail:          "smpn2jakarta@gmail.com",
			SchoolPhone:          "021-7777777",
			CommitteeCount:       16,
			CooperationLetterURL: "",
			IsActive:             true,
		},
		{
			Name:                 "SD Negeri 8 Bogor",
			Address:              "Jl. Pendidikan No. 8, Bogor",
			Latitude:             -6.6000,
			Longitude:            106.8000,
			ContactPerson:        "Ibu Sari",
			PhoneNumber:          "08288888888",
			StudentCount:         200,
			Category:             "SD",
			StudentCountGrade13:  105,
			StudentCountGrade46:  95,
			StaffCount:           22,
			NPSN:                 "20100008",
			PrincipalName:        "Sari Kusuma, S.Pd",
			SchoolEmail:          "sdn8bogor@gmail.com",
			SchoolPhone:          "0251-8888888",
			CommitteeCount:       13,
			CooperationLetterURL: "",
			IsActive:             true,
		},
	}

	for i := range schools {
		db.FirstOrCreate(&schools[i], models.School{Name: schools[i].Name})
	}

	log.Printf("Seeded %d schools\n", len(schools))
}

func seedEmployees(db *gorm.DB) {
	log.Println("Seeding employees...")

	var users []models.User
	db.Find(&users)

	positions := map[string]string{
		"kepala_sppg":       "Kepala SPPG",
		"kepala_yayasan":    "Kepala Yayasan",
		"akuntan":           "Akuntan",
		"ahli_gizi":         "Ahli Gizi",
		"pengadaan":         "Staff Pengadaan",
		"chef":              "Chef",
		"packing":           "Staff Packing",
		"driver":            "Driver",
		"asisten_lapangan":  "Asisten Lapangan",
		"kebersihan":        "Staff Kebersihan",
	}

	for _, user := range users {
		employee := models.Employee{
			UserID:      user.ID,
			NIK:         user.NIK,
			FullName:    user.FullName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Position:    positions[user.Role],
			JoinDate:    time.Now().AddDate(-rand.Intn(5)-1, 0, 0),
			IsActive:    true,
		}
		db.FirstOrCreate(&employee, models.Employee{UserID: user.ID})
	}

	log.Printf("Seeded %d employees\n", len(users))
}

func seedKitchenAssets(db *gorm.DB) {
	log.Println("Seeding kitchen assets...")

	assets := []models.KitchenAsset{
		{AssetCode: "AST-001", Name: "Kompor Gas Industri", Category: "Peralatan Masak", PurchaseDate: time.Now().AddDate(-2, 0, 0), PurchasePrice: 5000000, CurrentValue: 3500000, DepreciationRate: 15, Condition: "good", Location: "Dapur Utama"},
		{AssetCode: "AST-002", Name: "Oven Gas Besar", Category: "Peralatan Masak", PurchaseDate: time.Now().AddDate(-3, 0, 0), PurchasePrice: 8000000, CurrentValue: 4000000, DepreciationRate: 15, Condition: "good", Location: "Dapur Utama"},
		{AssetCode: "AST-003", Name: "Freezer Industrial", Category: "Pendingin", PurchaseDate: time.Now().AddDate(-1, 0, 0), PurchasePrice: 12000000, CurrentValue: 10000000, DepreciationRate: 10, Condition: "good", Location: "Gudang"},
		{AssetCode: "AST-004", Name: "Rice Cooker Kapasitas Besar", Category: "Peralatan Masak", PurchaseDate: time.Now().AddDate(-1, 0, 0), PurchasePrice: 3000000, CurrentValue: 2500000, DepreciationRate: 20, Condition: "good", Location: "Dapur Utama"},
		{AssetCode: "AST-005", Name: "Timbangan Digital", Category: "Alat Ukur", PurchaseDate: time.Now().AddDate(-4, 0, 0), PurchasePrice: 1500000, CurrentValue: 500000, DepreciationRate: 25, Condition: "fair", Location: "Gudang"},
	}

	for i := range assets {
		db.FirstOrCreate(&assets[i], models.KitchenAsset{AssetCode: assets[i].AssetCode})

		// Create maintenance record
		maintenance := models.AssetMaintenance{
			AssetID:         assets[i].ID,
			MaintenanceDate: time.Now().AddDate(0, -rand.Intn(6), 0),
			Description:     "Perawatan rutin dan pembersihan",
			Cost:            float64(rand.Intn(500000) + 100000),
			PerformedBy:     "Teknisi A",
		}
		db.Create(&maintenance)
	}

	log.Printf("Seeded %d kitchen assets\n", len(assets))
}

func seedCashFlowEntries(db *gorm.DB) {
	log.Println("Seeding cash flow entries...")

	categories := []string{"bahan_baku", "gaji", "utilitas", "operasional", "lainnya"}
	var kepalaSPPG models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)

	for i := 0; i < 20; i++ {
		entryType := "expense"
		if rand.Intn(10) == 0 {
			entryType = "income"
		}

		entry := models.CashFlowEntry{
			TransactionID: fmt.Sprintf("TRX-%06d", i+1),
			Date:          time.Now().AddDate(0, 0, -rand.Intn(90)),
			Category:      categories[rand.Intn(len(categories))],
			Type:          entryType,
			Amount:        float64(rand.Intn(10000000) + 1000000),
			Description:   fmt.Sprintf("Transaksi %d", i+1),
			Reference:     fmt.Sprintf("REF-%06d", i+1),
			CreatedBy:     kepalaSPPG.ID,
		}
		db.FirstOrCreate(&entry, models.CashFlowEntry{TransactionID: entry.TransactionID})
	}

	log.Println("Seeded 20 cash flow entries")
}

func seedBudgetTargets(db *gorm.DB) {
	log.Println("Seeding budget targets...")

	categories := []string{"bahan_baku", "gaji", "utilitas", "operasional", "lainnya"}
	year := time.Now().Year()

	for month := 1; month <= 12; month++ {
		for _, cat := range categories {
			target := models.BudgetTarget{
				Year:     year,
				Month:    month,
				Category: cat,
				Target:   float64(rand.Intn(50000000) + 10000000),
				Actual:   float64(rand.Intn(40000000) + 5000000),
			}
			db.FirstOrCreate(&target, models.BudgetTarget{Year: year, Month: month, Category: cat})
		}
	}

	log.Println("Seeded budget targets for 12 months")
}

func seedSystemConfig(db *gorm.DB) {
	log.Println("Seeding system config...")

	configs := []models.SystemConfig{
		{Key: "company_name", Value: "SPPG Pusat", DataType: "string", Category: "system"},
		{Key: "min_stock_alert", Value: "10", DataType: "int", Category: "inventory"},
		{Key: "max_menu_items", Value: "50", DataType: "int", Category: "system"},
		{Key: "delivery_radius", Value: "50", DataType: "float", Category: "logistics"},
	}

	var kepalaSPPG models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)

	for i := range configs {
		configs[i].UpdatedBy = kepalaSPPG.ID
		db.FirstOrCreate(&configs[i], models.SystemConfig{Key: configs[i].Key})
	}

	log.Printf("Seeded %d system configs\n", len(configs))
}

func seedMenuPlans(db *gorm.DB) {
	log.Println("Seeding menu plans...")

	var recipes []models.Recipe
	db.Find(&recipes)

	var kepalaSPPG, ahliGizi models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)
	db.Where("role = ?", "ahli_gizi").First(&ahliGizi)

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
	weekEnd := weekStart.AddDate(0, 0, 4)

	menuPlan := models.MenuPlan{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Status:    "approved",
		CreatedBy: ahliGizi.ID,
	}
	now2 := time.Now()
	menuPlan.ApprovedBy = &kepalaSPPG.ID
	menuPlan.ApprovedAt = &now2
	db.FirstOrCreate(&menuPlan, models.MenuPlan{WeekStart: weekStart})

	// Create menu items for each day
	weekdays := []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat"}
	for i, _ := range weekdays {
		date := weekStart.AddDate(0, 0, i)
		recipe := recipes[rand.Intn(len(recipes))]

		menuItem := models.MenuItem{
			MenuPlanID: menuPlan.ID,
			Date:       date,
			RecipeID:   recipe.ID,
			Portions:   rand.Intn(500) + 500,
		}
		db.FirstOrCreate(&menuItem, models.MenuItem{MenuPlanID: menuPlan.ID, Date: date})
	}

	log.Println("Seeded menu plan for current week")
}

func seedDeliveryTasks(db *gorm.DB) {
	log.Println("Seeding delivery tasks...")

	var schools []models.School
	db.Find(&schools)

	var drivers []models.User
	db.Where("role = ?", "driver").Find(&drivers)

	var recipes []models.Recipe
	db.Find(&recipes)

	statuses := []string{"pending", "in_progress", "completed"}
	today := time.Now()

	for i, school := range schools {
		driver := drivers[i%len(drivers)]
		status := statuses[rand.Intn(len(statuses))]

		task := models.DeliveryTask{
			TaskDate:   today,
			DriverID:   driver.ID,
			SchoolID:   school.ID,
			Portions:   school.StudentCount,
			Status:     status,
			RouteOrder: i + 1,
		}
		db.FirstOrCreate(&task, models.DeliveryTask{TaskDate: today, SchoolID: school.ID})

		// Add menu items to delivery
		for j := 0; j < rand.Intn(2)+1; j++ {
			recipe := recipes[rand.Intn(len(recipes))]
			menuItem := models.DeliveryMenuItem{
				DeliveryTaskID: task.ID,
				RecipeID:       recipe.ID,
				Portions:       task.Portions,
			}
			db.FirstOrCreate(&menuItem, models.DeliveryMenuItem{DeliveryTaskID: task.ID, RecipeID: recipe.ID})
		}

		// Create POD for completed deliveries
		if status == "completed" {
			pod := models.ElectronicPOD{
				DeliveryTaskID: task.ID,
				PhotoURL:       "https://storage.example.com/pod/pod_" + fmt.Sprintf("%d", task.ID) + ".jpg",
				SignatureURL:   "https://storage.example.com/sig/sig_" + fmt.Sprintf("%d", task.ID) + ".png",
				Latitude:       school.Latitude + (rand.Float64()-0.5)*0.001,
				Longitude:      school.Longitude + (rand.Float64()-0.5)*0.001,
				RecipientName:  school.ContactPerson,
				OmprengDropOff: school.StudentCount / 10,
				OmprengPickUp:  school.StudentCount/10 - rand.Intn(5),
				CompletedAt:    time.Now(),
			}
			db.FirstOrCreate(&pod, models.ElectronicPOD{DeliveryTaskID: task.ID})
		}
	}

	log.Printf("Seeded %d delivery tasks\n", len(schools))
}

func seedOmprengTracking(db *gorm.DB) {
	log.Println("Seeding ompreng tracking...")

	var schools []models.School
	db.Find(&schools)

	var kepalaSPPG models.User
	db.Where("role = ?", "kepala_sppg").First(&kepalaSPPG)

	// Create ompreng inventory
	ompInv := models.OmprengInventory{
		TotalOwned:    5000,
		AtKitchen:     2000,
		InCirculation: 2800,
		Missing:       200,
		LastUpdated:   time.Now(),
	}
	db.FirstOrCreate(&ompInv, models.OmprengInventory{ID: 1})

	// Create tracking records for each school
	for _, school := range schools {
		tracking := models.OmprengTracking{
			SchoolID:   school.ID,
			Date:       time.Now(),
			DropOff:    school.StudentCount / 10,
			PickUp:     school.StudentCount/10 - rand.Intn(5),
			Balance:    rand.Intn(50) + 10,
			RecordedBy: kepalaSPPG.ID,
		}
		db.FirstOrCreate(&tracking, models.OmprengTracking{SchoolID: school.ID, Date: time.Now()})
	}

	log.Println("Seeded ompreng tracking")
}

func seedNotifications(db *gorm.DB) {
	log.Println("Seeding notifications...")

	var users []models.User
	db.Find(&users)

	notificationTypes := []string{"low_stock", "po_approval", "delivery_complete", "system"}
	notificationTitles := []string{
		"Stok Bahan Menipis",
		"PO Menunggu Approval",
		"Pengiriman Selesai",
		"Pengumuman Sistem",
	}
	notificationMessages := []string{
		"Beberapa bahan baku telah mencapai batas minimum stok",
		"Terdapat Purchase Order yang memerlukan persetujuan",
		"Pengiriman ke sekolah telah berhasil diselesaikan",
		"Maintenance sistem akan dilakukan malam ini",
	}

	for i := 0; i < 15; i++ {
		user := users[rand.Intn(len(users))]
		notifType := notificationTypes[rand.Intn(len(notificationTypes))]
		idx := rand.Intn(len(notificationTitles))

		notif := models.Notification{
			UserID:    user.ID,
			Type:      notifType,
			Title:     notificationTitles[idx],
			Message:   notificationMessages[idx],
			IsRead:    rand.Intn(2) == 1,
			Link:      "/dashboard",
			CreatedAt: time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
		}
		db.Create(&notif)
	}

	log.Println("Seeded 15 notifications")
}

func seedAuditTrails(db *gorm.DB) {
	log.Println("Seeding audit trails...")

	var users []models.User
	db.Find(&users)

	actions := []string{"create", "update", "delete", "login", "approve"}
	entities := []string{"purchase_orders", "recipes", "ingredients", "delivery_tasks", "menu_plans"}

	for i := 0; i < 30; i++ {
		user := users[rand.Intn(len(users))]
		action := actions[rand.Intn(len(actions))]
		entity := entities[rand.Intn(len(entities))]

		audit := models.AuditTrail{
			UserID:    user.ID,
			Timestamp: time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour),
			Action:    action,
			Entity:    entity,
			EntityID:  fmt.Sprintf("%d", rand.Intn(100)+1),
			OldValue:  fmt.Sprintf(`{"status": "old"}`),
			NewValue:  fmt.Sprintf(`{"status": "new"}`),
			IPAddress: fmt.Sprintf("192.168.1.%d", rand.Intn(255)+1),
		}
		db.Create(&audit)
	}

	log.Println("Seeded 30 audit trails")
}

func seedWiFiConfig(db *gorm.DB) {
	log.Println("Seeding WiFi config...")

	wifis := []models.WiFiConfig{
		{SSID: "SPPG-Office", BSSID: "00:11:22:33:44:55", Location: "Kantor Pusat", IsActive: true},
		{SSID: "SPPG-Kitchen", BSSID: "00:11:22:33:44:66", Location: "Dapur Utama", IsActive: true},
		{SSID: "SPPG-Warehouse", BSSID: "00:11:22:33:44:77", Location: "Gudang", IsActive: true},
	}

	for i := range wifis {
		db.FirstOrCreate(&wifis[i], models.WiFiConfig{SSID: wifis[i].SSID})
	}

	// Create attendance records
	var employees []models.Employee
	db.Find(&employees)

	for _, emp := range employees {
		attendance := models.Attendance{
			EmployeeID: emp.ID,
			Date:       time.Now(),
			CheckIn:    time.Now().Add(-time.Duration(rand.Intn(3)+6) * time.Hour),
			WorkHours:  float64(rand.Intn(4) + 6),
			SSID:       wifis[rand.Intn(len(wifis))].SSID,
			BSSID:      wifis[rand.Intn(len(wifis))].BSSID,
		}
		db.FirstOrCreate(&attendance, models.Attendance{EmployeeID: emp.ID, Date: time.Now()})
	}

	log.Printf("Seeded WiFi config and attendance for %d employees\n", len(employees))
}
