package database

import (
	"fmt"
	"log"
	"time"

	"github.com/erp-sppg/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	// Configure GORM with optimizations
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// Disable foreign key constraints for better performance (handle in application)
		DisableForeignKeyConstraintWhenMigrating: true,
		// Prepare statements for better performance
		PrepareStmt: true,
		// Skip default transaction for single create, update, delete operations
		SkipDefaultTransaction: true,
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for optimal performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)                   // Maximum idle connections
	sqlDB.SetMaxOpenConns(100)                  // Maximum open connections
	sqlDB.SetConnMaxLifetime(time.Hour)         // Connection max lifetime
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)  // Connection max idle time

	log.Println("Database connection established with optimized settings")

	return db, nil
}
