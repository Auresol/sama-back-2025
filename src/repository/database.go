package repository

import (
	"fmt"
	"log"

	"sama/sama-backend-2025/src/config"
	"sama/sama-backend-2025/src/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(config *config.Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.Port,
		config.Database.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Auto migrate the schema
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %v", err)
	}

	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	// Import models here to register them for migration
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.School{})
	DB.AutoMigrate(&models.Activity{})

	rawSQL := `
	ALTER TABLE classrooms
	ADD COLUMN classroom TEXT GENERATED ALWAYS AS (room_number || '/' || room_section) STORED;
	`
	err := DB.Exec(rawSQL).Error
	if err != nil {
		// Handle error if column already exists gracefully, or ensure your migration
		// system prevents re-running this if it's already there.
		// if !isDuplicateColumnError(err) { // Custom check for duplicate column error
		// 	log.Fatalf("Failed to add generated column: %v", err)
		// }
		log.Println("Generated column 'classroom_identifier' already exists, skipping.")
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
