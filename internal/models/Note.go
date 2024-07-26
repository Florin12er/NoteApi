// Note.go
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Note struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"ID"`
	Title         string    `json:"title"`
	DashboardPath string    `json:"dashboard_path"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	LastChanged   time.Time `gorm:"autoUpdateTime" json:"last_changed"`
	LastRemove    time.Time `json:"last_removed"`
	UserID        uint      `json:"user_id"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (n *Note) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate will update the LastChanged time
func (n *Note) BeforeUpdate(tx *gorm.DB) error {
	n.LastChanged = time.Now()
	return nil
}

// SoftDelete updates the LastRemove time instead of deleting the record
func (n *Note) SoftDelete(tx *gorm.DB) error {
	n.LastRemove = time.Now()
	return tx.Save(n).Error
}

// Restore clears the LastRemove time to undelete the record
func (n *Note) Restore(tx *gorm.DB) error {
	n.LastRemove = time.Time{}
	return tx.Save(n).Error
}

