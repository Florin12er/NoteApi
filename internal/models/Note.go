// Note.go
package models

import (
	"gorm.io/gorm"
	"time"
)

type Note struct {
	gorm.Model
	Title         string    `json:"title"`
	DashboardPath string    `json:"dashboard_path"` // Store the path to the image
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	LastChanged   time.Time `json:"last_changed"`
	LastRemove    time.Time `json:"last_removed"`
	UserID        uint      `json:"user_id"`
}
