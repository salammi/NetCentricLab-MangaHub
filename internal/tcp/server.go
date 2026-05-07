// internal/tcp/server.go
package tcp

import (
	"encoding/json"
	"log"
	"net"
	"sync"
)

// ProgressUpdate matches the exact struct required by the project spec
type ProgressUpdate struct {
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Timestamp int64  `json:"timestamp"`
}

// ProgressSyncServer handles concurrent TCP connections and broadcasting
type ProgressSyncServer struct {
	Port        string
	Connections map[string]net.Conn
	Broadcast   chan ProgressUpdate
	mutex       sync.RWMutex // Protects the Connections map from concurrent read/write panics
}

func NewServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		Port:        port,
		Connections: make(map[string]net.Conn),
		Broadcast:   make(chan ProgressUpdate),
	}
}

func (s *ProgressSyncServer) Start() {
	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatalf("Failed to start TCP server on port %s: %v", s.Port, err)
	}
	defer listener.Close()

	log.Printf("TCP Sync Server listening on tcp://localhost:%s...", s.Port)

	// Launch a dedicated goroutine to listen for broadcast messages
	go s.handleBroadcasts()

	// Infinite loop to accept incoming client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept TCP connection: %v", err)
			continue
		}

		// Launch a new goroutine for every user that connects
		go s.handleConnection(conn)
	}
}

func (s *ProgressSyncServer) handleConnection(conn net.Conn) {
	// Use the remote IP address as a temporary connection ID
	clientID := conn.RemoteAddr().String()

	s.mutex.Lock()
	s.Connections[clientID] = conn
	s.mutex.Unlock()

	log.Printf("✓ New client connected to TCP Sync: %s", clientID)

	// Send an immediate welcome JSON message confirming connection
	welcomeMsg := []byte("{\"status\":\"connected\", \"message\":\"Welcome to MangaHub TCP Sync\"}\n")
	conn.Write(welcomeMsg)

	// Ensure the connection is cleaned up when the user disconnects
	defer func() {
		s.mutex.Lock()
		delete(s.Connections, clientID)
		s.mutex.Unlock()
		conn.Close()
		log.Printf("Client disconnected: %s", clientID)
	}()

	// Keep the connection open and listen for any direct TCP messages
	decoder := json.NewDecoder(conn)
	for {
		var update ProgressUpdate
		if err := decoder.Decode(&update); err != nil {
			break // Break the loop and disconnect if there's an error or EOF
		}

		// If a client pushes an update directly, forward it to the broadcast channel
		s.Broadcast <- update
	}
}

func (s *ProgressSyncServer) handleBroadcasts() {
	// This loop waits efficiently until a message enters the Broadcast channel
	for update := range s.Broadcast {
		msg, err := json.Marshal(update)
		if err != nil {
			continue
		}
		msg = append(msg, '\n') // Append a newline so the client knows the JSON string has ended

		s.mutex.RLock()
		for id, conn := range s.Connections {
			_, err := conn.Write(msg)
			if err != nil {
				log.Printf("Failed to send broadcast to %s: %v", id, err)
			}
		}
		s.mutex.RUnlock()

		log.Printf("Broadcasted update for Manga %s to %d connected clients", update.MangaID, len(s.Connections))
	}
}
