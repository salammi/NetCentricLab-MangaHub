// cmd/udp-server/main.go
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"NetCentricLab-MangaHub/internal/udp"
)

func main() {
	// Initialize the UDP Server on the default port 9091
	server := &udp.NotificationServer{
		Port:    "9091",
		Clients: make([]net.UDPAddr, 0),
	}

	// 1. FIXED: Removed the 'if err :=' wrapper because Start() does not return a value
	go func() {
		log.Println("Starting UDP Notification Server on port 9091...")
		server.Start() 
	}()

	// Simulate system-triggered chapter release notifications (UC-010)
	go func() {
		for {
			time.Sleep(30 * time.Second)
			log.Println("Simulating system broadcast...")
			server.Broadcast("one-piece", "One Piece Chapter 1096 is now available!")
		}
	}()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// The main thread will pause here until you press Ctrl+C
	<-sigChan

	log.Println("Shutting down UDP server...")
}