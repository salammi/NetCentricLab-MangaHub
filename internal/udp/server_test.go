// internal/udp/server_test.go
package udp

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestUDPNotificationSystem(t *testing.T) {
	server := NewServer("9092")
	server.Start()
	defer server.Stop()

	clientAddr, _ := net.ResolveUDPAddr("udp", "localhost:0")
	serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:9092")
	clientConn, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer clientConn.Close()

	// 1. Send subscription packet
	_, err = clientConn.WriteToUDP([]byte("subscribe"), serverAddr)
	if err != nil {
		t.Fatalf("Failed to send subscribe packet: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Wait for processing

    // 2. Consume the ACK packet to clear the buffer
    buffer := make([]byte, 1024)
    clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
    _, _, _ = clientConn.ReadFromUDP(buffer) 

	// 3. Verify registration
	server.mutex.RLock()
	if len(server.Clients) != 1 {
		t.Errorf("Expected 1 registered client, got %d", len(server.Clients))
	}
	server.mutex.RUnlock()

	// 4. Trigger Broadcast
	server.Broadcast("jujutsu-kaisen", "Chapter 248 is out!")

	// 5. Verify receipt
	clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, _, err := clientConn.ReadFromUDP(buffer)
	if err != nil {
		t.Fatalf("Client failed to receive broadcast: %v", err)
	}

	var notification Notification
	json.Unmarshal(buffer[:n], &notification)

	if notification.MangaID != "jujutsu-kaisen" {
		t.Errorf("Expected MangaID jujutsu-kaisen, got %s", notification.MangaID)
	}
}