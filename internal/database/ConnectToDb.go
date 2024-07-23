package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDb() {
	var err error
	dsn := os.Getenv("DB")
	
	// Create a new PostgreSQL configuration
	pgConfig := postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disables implicit prepared statement usage
	}

	// Open the database connection with the new configuration
	DB, err = gorm.Open(postgres.New(pgConfig), &gorm.Config{
		PrepareStmt: false, // Disable prepared statement caching
	})

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Set connection pool settings (optional, but recommended)
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
}

