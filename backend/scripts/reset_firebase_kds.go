package main

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Initialize Firebase with database URL
	config := &firebase.Config{
		DatabaseURL: "https://erp-sppg-default-rtdb.asia-southeast1.firebasedatabase.app",
	}
	opt := option.WithCredentialsFile("./firebase-credentials.json")
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	// Get database client
	client, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("Error getting database client: %v", err)
	}

	// Get today's date
	today := time.Now().Format("2006-01-02")

	// Delete cooking data for today
	cookingPath := fmt.Sprintf("/kds/cooking/%s", today)
	cookingRef := client.NewRef(cookingPath)
	if err := cookingRef.Delete(ctx); err != nil {
		log.Printf("Warning: Failed to delete cooking data: %v", err)
	} else {
		fmt.Printf("✓ Deleted cooking data at %s\n", cookingPath)
	}

	// Delete packing data for today
	packingPath := fmt.Sprintf("/kds/packing/%s", today)
	packingRef := client.NewRef(packingPath)
	if err := packingRef.Delete(ctx); err != nil {
		log.Printf("Warning: Failed to delete packing data: %v", err)
	} else {
		fmt.Printf("✓ Deleted packing data at %s\n", packingPath)
	}

	fmt.Println("\n✓ Firebase KDS data reset complete!")
}
