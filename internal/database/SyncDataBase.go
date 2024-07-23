package database

import (
	"NoteApi/internal/models"
	"log"
)

func SyncDatabase() {
	// Check if the users table exists
	if DB.Migrator().HasTable(&models.Note{}) {
		log.Println("Note table already exists. Migrating schema.")
		// AutoMigrate will only add missing columns and indexes, it won't delete/change existing columns
		if err := DB.AutoMigrate(&models.Note{}); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
	} else {
		// If the table doesn't exist, create it
		log.Println("Creating note table.")
		if err := DB.AutoMigrate(&models.Note{}); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
	}
}
