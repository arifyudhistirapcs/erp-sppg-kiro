package main

import (
	"fmt"
	"log"
	"time"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("=== Checking Drivers ===")
	fmt.Println()

	// Check for users with role 'driver'
	var drivers []models.User
	err = db.Where("role = ?", "driver").Find(&drivers).Error
	if err != nil {
		log.Fatalf("Failed to query drivers: %v", err)
	}

	fmt.Printf("Found %d drivers:\n", len(drivers))
	for _, driver := range drivers {
		fmt.Printf("- ID: %d, Name: %s, Email: %s, Active: %v\n", 
			driver.ID, driver.FullName, driver.Email, driver.IsActive)
	}
	fmt.Println()

	// Check for delivery records with status 'selesai_dipacking'
	var deliveryRecords []models.DeliveryRecord
	today := time.Now()
	err = db.Where("current_status = ?", "selesai_dipacking").
		Where("DATE(delivery_date) = DATE(?)", today).
		Preload("School").
		Preload("MenuItem").
		Preload("MenuItem.Recipe").
		Find(&deliveryRecords).Error
	if err != nil {
		log.Fatalf("Failed to query delivery records: %v", err)
	}

	fmt.Printf("Found %d delivery records with status 'selesai_dipacking' for today:\n", len(deliveryRecords))
	for _, record := range deliveryRecords {
		schoolName := "Unknown"
		if record.School.ID != 0 {
			schoolName = record.School.Name
		}
		menuName := "Unknown"
		if record.MenuItem.ID != 0 && record.MenuItem.Recipe.ID != 0 {
			menuName = record.MenuItem.Recipe.Name
		}
		fmt.Printf("- ID: %d, School: %s, Menu: %s, Portions: %d (Small: %d, Large: %d), Driver: %v\n",
			record.ID, schoolName, menuName, record.Portions, 
			record.PortionsSmall, record.PortionsLarge, record.DriverID)
	}
	fmt.Println()

	// If no drivers, suggest creating one
	if len(drivers) == 0 {
		fmt.Println("⚠️  No drivers found!")
		fmt.Println("You need to create users with role 'driver' in the database.")
		fmt.Println()
	}

	// If no delivery records, suggest the workflow
	if len(deliveryRecords) == 0 {
		fmt.Println("⚠️  No delivery records with status 'selesai_dipacking' found!")
		fmt.Println("Workflow to create delivery records:")
		fmt.Println("1. Create a menu plan and approve it")
		fmt.Println("2. Go to KDS Cooking and start cooking")
		fmt.Println("3. Mark cooking as complete")
		fmt.Println("4. Go to KDS Packing and start packing")
		fmt.Println("5. Mark packing as complete (this sets status to 'selesai_dipacking')")
		fmt.Println()
	}
}
