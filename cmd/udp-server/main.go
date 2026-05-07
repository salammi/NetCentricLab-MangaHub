// cmd/udp-server/main.go
package main

import (
	"NetCentricLab-MangaHub/internal/udp"
	"log"
)

func main() {
	log.Println("Initializing UDP Notification Server Component...")

	server := udp.NewServer("9091")
	server.Start()
}
