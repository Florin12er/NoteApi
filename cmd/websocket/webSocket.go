// websocket/websocket.go
package websocket

import (
    "NoteApi/internal/models"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "log"
    "net/http"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

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

    clients[ws] = true

    for {
        var msg Message
        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Printf("Error reading JSON: %v", err)
            delete(clients, ws)
            break
        }

        broadcast <- msg
    }
}

func HandleMessages() {
    for {
        msg := <-broadcast
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Printf("Error writing JSON: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}

func BroadcastNoteList(notes []models.Note) {
    noteList := make([]map[string]interface{}, len(notes))
    for i, note := range notes {
        noteList[i] = map[string]interface{}{
            "id":    note.ID,
            "title": note.Title,
        }
    }

    msg := Message{
        Type: "noteList",
        Data: noteList,
    }

    broadcast <- msg
}

func BroadcastNoteUpdate(note models.Note) {
    msg := Message{
        Type: "noteUpdate",
        Data: note,
    }

    broadcast <- msg
}

