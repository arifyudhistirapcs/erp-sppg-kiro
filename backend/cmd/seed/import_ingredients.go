package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Read CSV file
	csvPath := os.Args[1]
	if csvPath == "" {
		log.Fatal("Please provide CSV file path as argument")
	}

	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	// Skip header row
	records = records[1:]

	successCount := 0
	errorCount := 0

	for _, record := range records {
		if len(record) < 4 {
			log.Printf("Skipping invalid record: %v", record)
			errorCount++
			continue
		}

		no := record[0]
		name := strings.TrimSpace(record[1])
		category := strings.TrimSpace(record[2])
		unit := strings.TrimSpace(record[3])

		// Generate code based on number
		noInt, err := strconv.Atoi(no)
		if err != nil {
			log.Printf("Invalid number for %s: %v", name, err)
			errorCount++
			continue
		}
		code := fmt.Sprintf("B-%04d", noInt)

		// Create ingredient
		ingredient := models.Ingredient{
			Code:     code,
			Name:     name,
			Category: category,
			Unit:     unit,
		}

		if err := db.Create(&ingredient).Error; err != nil {
			log.Printf("Failed to create ingredient %s: %v", name, err)
			errorCount++
			continue
		}

		// Initialize inventory for this ingredient
		inventory := models.InventoryItem{
			IngredientID: ingredient.ID,
			Quantity:     0,
			MinThreshold: 100, // Default minimum threshold
		}

		if err := db.Create(&inventory).Error; err != nil {
			log.Printf("Failed to create inventory for %s: %v", name, err)
			errorCount++
			continue
		}

		successCount++
		if successCount%100 == 0 {
			log.Printf("Imported %d ingredients...", successCount)
		}
	}

	log.Printf("\n=== Import Complete ===")
	log.Printf("Success: %d", successCount)
	log.Printf("Errors: %d", errorCount)
	log.Printf("Total: %d", len(records))
}
