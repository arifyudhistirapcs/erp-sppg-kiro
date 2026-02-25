package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/firebase"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	fmt.Println("========================================")
	fmt.Println("Firebase Connection Test")
	fmt.Println("========================================")
	fmt.Println()

	// Initialize Firebase
	fmt.Println("1. Initializing Firebase...")
	fmt.Printf("   Credentials Path: %s\n", cfg.FirebaseCredentialsPath)
	fmt.Printf("   Database URL: %s\n", cfg.FirebaseDatabaseURL)
	fmt.Println()

	app, err := firebase.Initialize(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firebase: %v", err)
	}
	fmt.Println("✅ Firebase initialized successfully!")
	fmt.Println()

	// Test Realtime Database connection
	fmt.Println("2. Testing Realtime Database connection...")
	ctx := context.Background()
	
	dbClient, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get database client: %v", err)
	}
	fmt.Println("✅ Database client created successfully!")
	fmt.Println()

	// Test write operation
	fmt.Println("3. Testing write operation...")
	testData := map[string]interface{}{
		"test_timestamp": time.Now().Unix(),
		"test_message":   "Firebase connection test successful",
		"version":        "1.0.0",
	}

	testRef := dbClient.NewRef("/test/connection")
	if err := testRef.Set(ctx, testData); err != nil {
		log.Fatalf("❌ Failed to write test data: %v", err)
	}
	fmt.Println("✅ Write operation successful!")
	fmt.Println()

	// Test read operation
	fmt.Println("4. Testing read operation...")
	var readData map[string]interface{}
	if err := testRef.Get(ctx, &readData); err != nil {
		log.Fatalf("❌ Failed to read test data: %v", err)
	}
	fmt.Println("✅ Read operation successful!")
	fmt.Printf("   Data: %+v\n", readData)
	fmt.Println()

	// Test dashboard path
	fmt.Println("5. Testing dashboard path...")
	dashboardRef := dbClient.NewRef("/dashboard/test")
	dashboardData := map[string]interface{}{
		"test_time": time.Now().Format(time.RFC3339),
		"status":    "connected",
	}
	if err := dashboardRef.Set(ctx, dashboardData); err != nil {
		log.Fatalf("❌ Failed to write dashboard test data: %v", err)
	}
	fmt.Println("✅ Dashboard path accessible!")
	fmt.Println()

	// Test KDS path
	fmt.Println("6. Testing KDS path...")
	today := time.Now().Format("2006-01-02")
	kdsRef := dbClient.NewRef(fmt.Sprintf("/kds/cooking/%s/test", today))
	kdsData := map[string]interface{}{
		"recipe_id": 1,
		"status":    "pending",
		"timestamp": time.Now().Unix(),
	}
	if err := kdsRef.Set(ctx, kdsData); err != nil {
		log.Fatalf("❌ Failed to write KDS test data: %v", err)
	}
	fmt.Println("✅ KDS path accessible!")
	fmt.Println()

	// Summary
	fmt.Println("========================================")
	fmt.Println("✅ All Firebase tests passed!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Your Firebase is configured correctly and ready to use.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Start the backend server: go run cmd/server/main.go")
	fmt.Println("2. Test the dashboard API")
	fmt.Println("3. Check Firebase Console to see the test data")
	fmt.Println()
}
