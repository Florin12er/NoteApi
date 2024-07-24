// internal/handlers/notes_handlers.go

package handlers

import (
    "NoteApi/internal/database"
    "NoteApi/internal/models"
    "NoteApi/cmd/websocket"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

func CreateNote(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
        return
    }

    userIDUint, ok := userID.(uint)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
        return
    }

    var note models.Note
    if err := c.ShouldBindJSON(&note); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    note.UserID = userIDUint
    note.CreatedAt = time.Now()
    note.LastChanged = time.Now()

    if err := database.DB.Create(&note).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
        return
    }

    // Broadcast the new note to the user
    websocket.BroadcastNoteUpdateToUser(note, userIDUint)

    // Broadcast the updated note list to the user
    var notes []models.Note
    database.DB.Where("user_id = ?", userIDUint).Select("id, title content").Find(&notes)
    websocket.BroadcastNoteListToUser(notes, userIDUint)

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
    userID, _ := c.Get("user_id")
    id := c.Param("id")
    var note models.Note

    if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
        return
    }

    if err := c.ShouldBindJSON(&note); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    note.LastChanged = time.Now()

    if err := database.DB.Save(&note).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
        return
    }

    // Broadcast the updated note to the user
    websocket.BroadcastNoteUpdateToUser(note, userID.(uint))

    // Broadcast the updated note list to the user
    var notes []models.Note
    database.DB.Where("user_id = ?", userID).Select("id, title").Find(&notes)
    websocket.BroadcastNoteListToUser(notes, userID.(uint))

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
    websocket.BroadcastNoteDeleteToUser(note.ID, userID.(uint))

    // Broadcast the updated note list to the user
    var notes []models.Note
    database.DB.Where("user_id = ?", userID).Select("id, title").Find(&notes)
    websocket.BroadcastNoteListToUser(notes, userID.(uint))

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

