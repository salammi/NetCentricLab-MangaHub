// internal/websocket/hub.go
package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type ChatHub struct {
	Clients    map[*websocket.Conn]string
	Broadcast  chan ChatMessage
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
}

func NewHub() *ChatHub {
	return &ChatHub{
		Clients:    make(map[*websocket.Conn]string),
		Broadcast:  make(chan ChatMessage),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = "anonymous"
			log.Println("New WebSocket client connected")

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
				log.Println("WebSocket client disconnected")
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("Error broadcasting: %v", err)
					client.Close()
					delete(h.Clients, client)
				}
			}
		}
	}
}