// internal/handlers/notes_handlers.go

package handlers

import (
	"NoteApi/cmd/websocket"
	"NoteApi/internal/database"
	"NoteApi/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func CreateNote(c *gin.Context) {
	// Parse the multipart form
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var userIDUUID uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		userIDUUID = v
	case string:
		var err error
		userIDUUID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	note := models.Note{
		UserID:  userIDUUID,
		Title:   c.Request.FormValue("title"),
		Content: c.Request.FormValue("content"),
	}

	// Handle file upload
	file, header, err := c.Request.FormFile("dashboard_image")
	if err == nil {
		// Generate a unique filename
		fileExt := filepath.Ext(header.Filename)
		newFilename := uuid.New().String() + fileExt

		// Create the uploads directory if it doesn't exist
		uploadsDir := "./uploads"
		if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "Failed to create uploads directory"},
			)
			return
		}

		// Create the file
		dst, err := os.Create(filepath.Join(uploadsDir, newFilename))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the file"})
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
			return
		}

		note.DashboardPath = filepath.Join("uploads", newFilename)
		defer file.Close()
	}

	if err := database.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	// Broadcast the new note to the user
	websocket.BroadcastNoteUpdateToUser(note, userIDUUID)

	// Broadcast the updated note list to the user
	var notes []models.Note
	database.DB.Where("user_id = ?", userIDUUID).
		Select("id, title, content, dashboard_path").
		Find(&notes)
	websocket.BroadcastNoteListToUser(notes, userIDUUID)

	c.JSON(http.StatusCreated, note)
}

func GetNote(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id := c.Param("id")
	var note models.Note

	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func UpdateNote(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var userIDUUID uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		userIDUUID = v
	case string:
		var err error
		userIDUUID, err = uuid.Parse(v)
		if err != nil {
			log.Printf("Invalid user ID format: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}
	default:
		log.Println("Invalid user ID type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	id := c.Param("id")
	var note models.Note

	if err := database.DB.Where("id = ? AND user_id = ?", id, userIDUUID).First(&note).Error; err != nil {
		log.Printf("Note not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Parse the multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		log.Printf("Failed to parse form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Update note fields
	note.Title = c.Request.FormValue("title")
	note.Content = c.Request.FormValue("content")

	// Handle file upload
	file, header, err := c.Request.FormFile("dashboard_image")
	if err == nil {
		// Generate a unique filename
		fileExt := filepath.Ext(header.Filename)
		newFilename := uuid.New().String() + fileExt

		// Create the uploads directory if it doesn't exist
		uploadsDir := "./uploads"
		if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
			log.Printf("Failed to create uploads directory: %v", err)
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "Failed to create uploads directory"},
			)
			return
		}

		// Create the file
		dst, err := os.Create(filepath.Join(uploadsDir, newFilename))
		if err != nil {
			log.Printf("Failed to create the file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the file"})
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("Failed to save the file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
			return
		}

		note.DashboardPath = filepath.Join("uploads", newFilename)
		defer file.Close()
	} else if err != http.ErrMissingFile {
		log.Printf("Failed to handle file upload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle file upload"})
		return
	}

	if err := database.DB.Save(&note).Error; err != nil {
		log.Printf("Failed to update note: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	// Broadcast the updated note to the user
	websocket.BroadcastNoteUpdateToUser(note, userIDUUID)

	// Broadcast the updated note list to the user
	var notes []models.Note
	database.DB.Where("user_id = ?", userIDUUID).
		Select("id, title, content, dashboard_path").
		Find(&notes)
	websocket.BroadcastNoteListToUser(notes, userIDUUID)

	c.JSON(http.StatusOK, note)
}

func DeleteNote(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id := c.Param("id")
	var note models.Note

	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if err := database.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	// Broadcast the deleted note ID to the user
	websocket.BroadcastNoteDeleteToUser(note.ID, userID.(uuid.UUID))

	// Broadcast the updated note list to the user
	var notes []models.Note
	database.DB.Where("user_id = ?", userID).Select("id, title").Find(&notes)
	websocket.BroadcastNoteListToUser(notes, userID.(uuid.UUID))

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func ListNotes(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var notes []models.Note

	if err := database.DB.Where("user_id = ?", userID).Select("id, title, content, dashboard_path").Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes"})
		return
	}

	c.JSON(http.StatusOK, notes)
}
