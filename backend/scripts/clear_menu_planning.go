package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/firebase"
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
		os.Exit(1)
	}

	// Initialize Firebase
	firebaseApp, err := firebase.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()
	dbClient, err := firebaseApp.Database(ctx)
	if err != nil {
		log.Fatalf("Failed to get Firebase database client: %v", err)
		os.Exit(1)
	}

	fmt.Println("=== Clearing Menu Planning Data ===")
	fmt.Println()

	// Clear PostgreSQL data
	fmt.Println("Clearing PostgreSQL data...")
	
	// Delete in correct order to respect foreign key constraints
	tables := []string{
		"status_transitions",
		"delivery_records",
		"menu_item_school_allocations",
		"menu_items",
		"menu_plans",
	}

	for _, table := range tables {
		result := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if result.Error != nil {
			log.Printf("Warning: Failed to clear %s: %v", table, result.Error)
		} else {
			fmt.Printf("✓ Cleared %s (%d rows deleted)\n", table, result.RowsAffected)
		}
	}

	// Reset sequences
	fmt.Println("\nResetting sequences...")
	sequences := []string{
		"menu_plans",
		"menu_items",
		"menu_item_school_allocations",
		"delivery_records",
		"status_transitions",
	}

	for _, seq := range sequences {
		result := db.Exec(fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", seq))
		if result.Error != nil {
			log.Printf("Warning: Failed to reset sequence for %s: %v", seq, result.Error)
		} else {
			fmt.Printf("✓ Reset sequence for %s\n", seq)
		}
	}

	// Clear Firebase KDS data
	fmt.Println("\nClearing Firebase KDS data...")
	
	// Clear cooking data
	cookingRef := dbClient.NewRef("/kds/cooking")
	if err := cookingRef.Delete(ctx); err != nil {
		log.Printf("Warning: Failed to clear Firebase cooking data: %v", err)
	} else {
		fmt.Println("✓ Cleared Firebase cooking data")
	}

	// Clear packing data
	packingRef := dbClient.NewRef("/kds/packing")
	if err := packingRef.Delete(ctx); err != nil {
		log.Printf("Warning: Failed to clear Firebase packing data: %v", err)
	} else {
		fmt.Println("✓ Cleared Firebase packing data")
	}

	// Verify deletion
	fmt.Println("\n=== Verification ===")
	type Count struct {
		Count int64
	}

	for _, table := range tables {
		var count Count
		db.Raw(fmt.Sprintf("SELECT COUNT(*) as count FROM %s", table)).Scan(&count)
		fmt.Printf("%s: %d rows\n", table, count.Count)
	}

	fmt.Println("\n✓ Menu planning data cleared successfully!")
}
