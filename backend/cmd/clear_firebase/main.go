package main

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	// Initialize Firebase
	opt := option.WithCredentialsFile("../../firebase-credentials.json")
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		DatabaseURL: os.Getenv("FIREBASE_DATABASE_URL"),
	}, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	// Get database client
	client, err := app.Database(context.Background())
	if err != nil {
		log.Fatalf("Error getting database client: %v", err)
	}

	ctx := context.Background()

	// Clear KDS data
	paths := []string{
		"/kds/cooking",
		"/kds/packing",
		"/delivery",
		"/monitoring",
		"/cleaning/pending",
	}

	fmt.Println("🔥 Clearing Firebase KDS data...")
	fmt.Println()

	for _, path := range paths {
		fmt.Printf("Clearing %s...\n", path)
		if err := client.NewRef(path).Delete(ctx); err != nil {
			log.Printf("⚠️  Warning: Failed to clear %s: %v\n", path, err)
		} else {
			fmt.Printf("✓ Cleared %s\n", path)
		}
	}

	fmt.Println()
	fmt.Println("✅ All Firebase KDS data cleared successfully!")
}
