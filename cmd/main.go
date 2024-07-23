// main.go
package main

import (
	"NoteApi/cmd/websocket"
	"NoteApi/internal/database"
	"NoteApi/internal/handlers"
	"NoteApi/internal/middleware"
	"NoteApi/pkg/utils"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// In your main function, before defining routes
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // or whatever your frontend URL is
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Note routes
	r.POST("/notes", middleware.CheckAuthenticated(), handlers.CreateNote)
	r.GET("/notes/:id", middleware.CheckAuthenticated(), handlers.GetNote)
	r.PUT("/notes/:id", middleware.CheckAuthenticated(), handlers.UpdateNote)
	r.DELETE("/notes/:id", middleware.CheckAuthenticated(), handlers.DeleteNote)
	r.GET("/notes", middleware.CheckAuthenticated(), handlers.ListNotes)

	// Image upload route
	r.POST("/upload", middleware.CheckAuthenticated(), handlers.UploadImage)

	// WebSocket route
	r.GET("/ws", func(c *gin.Context) {
		websocket.HandleConnections(c)
	})

	go websocket.HandleMessages()

	// Start the server
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
