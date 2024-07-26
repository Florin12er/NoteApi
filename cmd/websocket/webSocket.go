package websocket

import (
	"NoteApi/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
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
	userID uint
}

var clients = make(map[*Client]bool)
var broadcast = make(chan Message)
var mu sync.Mutex

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func HandleConnections(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	userID, exists := c.Get("user_id") // Changed from "userID" to "user_id" to match the middleware
	if !exists {
		log.Printf("User ID not found in context")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		log.Printf("User ID is not of type uint")
		return
	}

	client := &Client{
		conn:   ws,
		userID: userIDUint,
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

func BroadcastNoteListToUser(notes []models.Note, userID uint) {
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


func BroadcastNoteUpdateToUser(note models.Note, userID uint) {
	msg := Message{
		Type: "noteUpdate",
		Data: map[string]interface{}{
			"id":      note.ID.String(), // Convert UUID to string
			"title":   note.Title,
			"content": note.Content,
            "dashboard_path": note.DashboardPath,
		},
	}

	broadcastToUser(msg, userID)
}

func BroadcastNoteDeleteToUser(noteID uuid.UUID, userID uint) {
	msg := Message{
		Type: "noteDelete",
		Data: noteID.String(), // Convert UUID to string
	}

	broadcastToUser(msg, userID)
}

func broadcastToUser(msg Message, userID uint) {
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
