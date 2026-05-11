// internal/udp/server.go
package udp

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type NotificationServer struct {
	Port    string
	Clients []net.UDPAddr
	mutex   sync.RWMutex
	conn    *net.UDPConn
    quit    chan struct{}
}

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
        quit:    make(chan struct{}),
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

	log.Printf("UDP Notification Server listening on udp://localhost:%s...", s.Port)
    go s.listen()
}

func (s *NotificationServer) listen() {
	buffer := make([]byte, 1024)
	for {
        select {
        case <-s.quit:
            return
        default:
            n, clientAddr, err := s.conn.ReadFromUDP(buffer)
            if err != nil {
                if strings.Contains(err.Error(), "use of closed network connection") {
                    break // Stop looping if connection is closed
                }
                continue
            }

            msg := strings.TrimSpace(string(buffer[:n]))
            if msg == "subscribe" || strings.Contains(msg, "subscribe") {
                s.registerClient(*clientAddr)
            } else if msg == "unsubscribe" || strings.Contains(msg, "unsubscribe") {
                s.unregisterClient(*clientAddr)
            }
        }
	}
}

func (s *NotificationServer) registerClient(addr net.UDPAddr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, client := range s.Clients {
		if client.String() == addr.String() {
			return
		}
	}
	s.Clients = append(s.Clients, addr)
	log.Printf("✓ New UDP client registered for notifications: %s", addr.String())
	s.conn.WriteToUDP([]byte(`{"status":"subscribed"}`+"\n"), &addr)
}

func (s *NotificationServer) unregisterClient(addr net.UDPAddr) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    for i, client := range s.Clients {
        if client.String() == addr.String() {
            s.Clients = append(s.Clients[:i], s.Clients[i+1:]...)
            log.Printf("✓ UDP client unregistered: %s", addr.String())
            return
        }
    }
}

func (s *NotificationServer) Broadcast(mangaID string, message string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.conn == nil || len(s.Clients) == 0 {
		return
	}

    notification := Notification{
        Type: "new_chapter",
        MangaID: mangaID,
        Message: message,
        Timestamp: time.Now().Unix(),
    }

	data, err := json.Marshal(notification)
	if err != nil {
		return
	}
	data = append(data, '\n')

	for _, client := range s.Clients {
		_, err := s.conn.WriteToUDP(data, &client)
        if err != nil {
            log.Printf("Failed to broadcast to %s: %v", client.String(), err)
        }
	}
	log.Printf("Broadcasted chapter notification to %d UDP clients", len(s.Clients))
}

func (s *NotificationServer) Stop() {
    close(s.quit)
    if s.conn != nil {
        s.conn.Close()
    }
}