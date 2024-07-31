package websocket

import (
	"NoteApi/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" // Import JWT library
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn   *websocket.Conn
	userID uuid.UUID
}

var clients = make(map[*Client]bool)
var broadcast = make(chan Message)
var mu sync.Mutex

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func HandleConnections(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		log.Println("No token provided")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	// Validate token
	claims := &jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	if err != nil {
		log.Printf("Invalid token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	if !parsedToken.Valid {
		log.Println("Token is not valid")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Extract user ID from claims
	userIDStr, ok := (*claims)["sub"].(string)
	if !ok {
		log.Println("User ID not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "https://note-taking-dusky.vercel.app" // Replace with your frontend URL
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	log.Printf("WebSocket connection established for user: %s", userID)

	client := &Client{
		conn:   ws,
		userID: userID,
	}

	mu.Lock()
	clients[client] = true
	mu.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			mu.Lock()
			delete(clients, client)
			mu.Unlock()
			break
		}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		mu.Lock()
		for client := range clients {
			err := client.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func BroadcastNoteListToUser(notes []models.Note, userID uuid.UUID) {
	noteList := make([]map[string]interface{}, len(notes))
	for i, note := range notes {
		noteList[i] = map[string]interface{}{
			"ID":      note.ID.String(), // Convert UUID to string
			"title":   note.Title,
			"content": note.Content,
			// Add other fields as needed
		}
	}

	msg := Message{
		Type: "noteList",
		Data: noteList,
	}

	broadcastToUser(msg, userID)
}

func BroadcastNoteUpdateToUser(note models.Note, userID uuid.UUID) {
	msg := Message{
		Type: "noteUpdate",
		Data: map[string]interface{}{
			"id":             note.ID.String(), // Convert UUID to string
			"title":          note.Title,
			"content":        note.Content,
			"dashboard_path": note.DashboardPath,
		},
	}

	broadcastToUser(msg, userID)
}

func BroadcastNoteDeleteToUser(noteID uuid.UUID, userID uuid.UUID) {
	msg := Message{
		Type: "noteDelete",
		Data: noteID.String(), // Convert UUID to string
	}

	broadcastToUser(msg, userID)
}

func broadcastToUser(msg Message, userID uuid.UUID) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		if client.userID == userID {
			err := client.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
	}
}
