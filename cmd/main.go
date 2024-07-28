// main.go
package main

import (
    "NoteApi/cmd/websocket"
    "NoteApi/internal/database"
    "NoteApi/internal/handlers"
    "NoteApi/internal/middleware"
    "NoteApi/pkg/utils"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
    "time"
)

func init() {
    utils.LoadEnv()
    database.ConnectToDb()
    database.SyncDatabase()
}

func main() {
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
        panic(err)
    }

    config := cors.Config{
        AllowOrigins:     []string{"https://note-taking-dusky.vercel.app"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type", "Set-Cookie", "Authorization"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    r.Use(cors.New(config))

    // Serve static files from the uploads directory
    r.Static("/uploads", "../cmd/uploads")

    // Health check route
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Note routes
    r.POST("/notes", middleware.CheckAuthenticated(), handlers.CreateNote)
    r.GET("/notes/:id", middleware.CheckAuthenticated(), handlers.GetNote)
    r.PUT("/notes/:id", middleware.CheckAuthenticated(), handlers.UpdateNote)
    r.DELETE("/notes/:id", middleware.CheckAuthenticated(), handlers.DeleteNote)
    r.GET("/notes", middleware.CheckAuthenticated(), handlers.ListNotes)

    // Image upload route
    r.POST("/upload", middleware.CheckAuthenticated(), handlers.UploadImage)

    // WebSocket route
    r.GET("/ws", middleware.CheckAuthenticated(), func(c *gin.Context) {
        websocket.HandleConnections(c)
    })

    go websocket.HandleMessages()

    // Ensure the uploads directory exists
    if err := utils.EnsureDir("uploads"); err != nil {
        log.Fatalf("failed to create uploads directory: %v", err)
    }

    // Start the server
    if err := r.Run(); err != nil {
        log.Fatalf("failed to run server: %v", err)
    }
}

