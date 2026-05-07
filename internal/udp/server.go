// internal/udp/server.go
package udp

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
)

// NotificationServer matches the exact struct required by the project spec
type NotificationServer struct {
	Port    string
	Clients []net.UDPAddr
	mutex   sync.RWMutex
	conn    *net.UDPConn
}

// Notification matches the exact struct required by the project spec
type Notification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func NewServer(port string) *NotificationServer {
	return &NotificationServer{
		Port:    port,
		Clients: make([]net.UDPAddr, 0),
	}
}

func (s *NotificationServer) Start() {
	addr, err := net.ResolveUDPAddr("udp", ":"+s.Port)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to start UDP server on port %s: %v", s.Port, err)
	}
	s.conn = conn
	defer conn.Close()

	log.Printf("UDP Notification Server listening on udp://localhost:%s...", s.Port)

	buffer := make([]byte, 1024)
	for {
		// Connectionless listening: Wait for any packet from any address
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		msg := strings.TrimSpace(string(buffer[:n]))

		// UC-009: Register for UDP Notifications
		if msg == "subscribe" {
			s.registerClient(*clientAddr)
		}
	}
}

func (s *NotificationServer) registerClient(addr net.UDPAddr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Prevent duplicate registrations
	for _, client := range s.Clients {
		if client.String() == addr.String() {
			return
		}
	}

	s.Clients = append(s.Clients, addr)
	log.Printf("✓ New UDP client registered for notifications: %s", addr.String())

	// Send confirmation packet back to the client
	s.conn.WriteToUDP([]byte("Successfully subscribed to chapter notifications!\n"), &addr)
}

// UC-010: Send Chapter Release Notification
func (s *NotificationServer) Broadcast(notification Notification) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.conn == nil || len(s.Clients) == 0 {
		return // No one to notify
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return
	}
	data = append(data, '\n')

	// Fire out the notification to every registered IP address
	for _, client := range s.Clients {
		s.conn.WriteToUDP(data, &client)
	}
	log.Printf("Broadcasted chapter notification to %d UDP clients", len(s.Clients))
}
