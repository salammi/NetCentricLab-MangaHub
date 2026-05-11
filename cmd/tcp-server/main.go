// internal/tcp/server.go
package tcp

import (
	"encoding/json"
	"log"
	"net"
	"sync"
)

type ProgressSyncServer struct {
	port        string
	Connections map[string]net.Conn
	mutex       sync.Mutex
}

// Renamed to NewTCPServer to match standard conventions
func NewTCPServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		port:        port,
		Connections: make(map[string]net.Conn),
	}
}

func (s *ProgressSyncServer) Start() {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}
	defer lis.Close()
	
	log.Printf("📡 TCP Sync Server listening on :%s", s.port)

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("TCP Accept Error:", err)
			continue
		}
		// Handle each connection in a new goroutine
		go s.handleConnection(conn)
	}
}

func (s *ProgressSyncServer) handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	
	s.mutex.Lock()
	s.Connections[addr] = conn // Save connection to keep it alive
	s.mutex.Unlock()

	log.Printf("New TCP client connected: %s", addr)

	// Listen for disconnects
	buffer := make([]byte, 1024)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			break // Client disconnected
		}
	}

	s.mutex.Lock()
	delete(s.Connections, addr)
	s.mutex.Unlock()
	conn.Close()
	log.Printf("TCP client disconnected: %s", addr)
}

// Broadcast sends progress updates to all connected TCP clients
func (s *ProgressSyncServer) Broadcast(data interface{}) {
	msg, _ := json.Marshal(data)
	msg = append(msg, '\n') // Add newline so bufio.Scanner can read it
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for addr, conn := range s.Connections {
		_, err := conn.Write(msg)
		if err != nil {
			log.Printf("Failed to broadcast to %s, dropping connection", addr)
			conn.Close()
			delete(s.Connections, addr)
		}
	}
}