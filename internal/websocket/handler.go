// internal/websocket/handler.go
package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for the sake of the academic demo
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

// ServeWs upgrades the HTTP connection to a WebSocket
func ServeWs(hub *ChatHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket Upgrade Error:", err)
			return
		}

		// Register the new client into the hub
		hub.Register <- conn

		// Spin up a goroutine to constantly read messages from this client
		go func() {
			defer func() {
				hub.Unregister <- conn
			}()

			for {
				var msg ChatMessage
				err := conn.ReadJSON(&msg)
				if err != nil {
					log.Println("Client disconnected or sent invalid JSON")
					break
				}
				// Send the message to the hub to be broadcasted to everyone
				hub.Broadcast <- msg
			}
		}()
	}
}