package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "net/http"
    "path/filepath"
    "strings"
)

const (
    MaxUploadSize = 10 << 20 // 10 MB
    UploadPath    = "uploads"
)

func UploadImage(c *gin.Context) {
    // Set a lower memory limit for multipart forms (default is 32 MiB)
    c.Request.ParseMultipartForm(MaxUploadSize)

    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()

    // Validate file size
    if header.Size > MaxUploadSize {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File is too large"})
        return
    }

    // Validate file type
    if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File is not an image"})
        return
    }

    // Generate a unique filename
    ext := filepath.Ext(header.Filename)
    newFilename := uuid.New().String() + ext
    dst := filepath.Join(UploadPath, newFilename)

    if err := c.SaveUploadedFile(header, dst); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }

    // Return the path to be stored in the Note model
    c.JSON(http.StatusOK, gin.H{"dashboard_path": dst})
}

