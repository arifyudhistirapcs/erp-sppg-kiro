package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/erp-sppg/backend/internal/cache"
	"github.com/erp-sppg/backend/internal/config"
	"github.com/erp-sppg/backend/internal/database"
	"github.com/erp-sppg/backend/internal/firebase"
	"github.com/erp-sppg/backend/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Redis cache (optional)
	var cacheService *cache.CacheService
	if cfg.EnableCache {
		redisCache, err := cache.NewRedisCache(cache.CacheConfig{
			Host:     cfg.RedisHost,
			Port:     cfg.RedisPort,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize Redis cache: %v", err)
			log.Println("Continuing without cache...")
		} else {
			cacheService = cache.NewCacheService(redisCache)
			log.Println("Redis cache initialized successfully")
		}
	}

	// Initialize Firebase
	firebaseApp, err := firebase.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Start database performance monitoring
	perfMonitor := database.NewPerformanceMonitor(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go perfMonitor.StartPerformanceMonitoring(ctx, 5*time.Minute)

	// Setup Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize router
	r := router.Setup(db, firebaseApp, cfg, cacheService)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
