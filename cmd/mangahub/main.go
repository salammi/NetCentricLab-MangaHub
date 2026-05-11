// cmd/mangahub/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "notify" || os.Args[2] != "subscribe" {
		fmt.Println("Usage: mangahub notify subscribe")
		return
	}

	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:9091")
	if err != nil {
		log.Fatalf("Error resolving server address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("subscribe"))
	if err != nil {
		log.Fatalf("Error sending subscription: %v", err)
	}

	fmt.Println("✓ Successfully connected to UDP Notification System")
	fmt.Println("Listening for chapter releases...")

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}
        
        msg := string(buffer[:n])
        // Ignore the ACK packet
        if msg != "{\"status\":\"subscribed\"}\n" {
		    fmt.Printf("\n[NEW RELEASE!] %s", msg)
        }
	}
}