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
    var note models.Note
    if err := c.ShouldBindJSON(&note); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    note.CreatedAt = time.Now()
    note.LastChanged = time.Now()

    if err := database.DB.Create(&note).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
        return
    }

    // Broadcast the new note
    websocket.BroadcastNoteUpdate(note)

    // Broadcast the updated note list
    var notes []models.Note
    database.DB.Select("id, title").Find(&notes)
    websocket.BroadcastNoteList(notes)

    c.JSON(http.StatusCreated, note)
}

func GetNote(c *gin.Context) {
    id := c.Param("id")
    var note models.Note

    if err := database.DB.First(&note, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
        return
    }

    c.JSON(http.StatusOK, note)
}

func UpdateNote(c *gin.Context) {
    id := c.Param("id")
    var note models.Note

    if err := database.DB.First(&note, id).Error; err != nil {
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

    // Broadcast the updated note
    websocket.BroadcastNoteUpdate(note)

    // Broadcast the updated note list
    var notes []models.Note
    database.DB.Select("id, title").Find(&notes)
    websocket.BroadcastNoteList(notes)

    c.JSON(http.StatusOK, note)
}

func DeleteNote(c *gin.Context) {
    id := c.Param("id")
    var note models.Note

    if err := database.DB.First(&note, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
        return
    }

    if err := database.DB.Delete(&note).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
        return
    }

    // Broadcast the deleted note ID
    websocket.BroadcastNoteDelete(note.ID)

    // Broadcast the updated note list
    var notes []models.Note
    database.DB.Select("id, title").Find(&notes)
    websocket.BroadcastNoteList(notes)

    c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func ListNotes(c *gin.Context) {
    var notes []models.Note

    if err := database.DB.Select("id, title, content, dashboard_path").Find(&notes).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes"})
        return
    }

    c.JSON(http.StatusOK, notes)
}

