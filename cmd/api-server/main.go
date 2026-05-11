// cmd/api-server/main.go
package main

import (
	"database/sql"
	"log"
	"net"

	"NetCentricLab-MangaHub/internal/auth"
	"NetCentricLab-MangaHub/internal/manga"
	"NetCentricLab-MangaHub/internal/tcp"
	"NetCentricLab-MangaHub/internal/udp"
	"NetCentricLab-MangaHub/internal/websocket"
	"NetCentricLab-MangaHub/pkg/database"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	Router    *gin.Engine
	Database  *sql.DB
	JWTSecret string
}

func main() {
	log.Println("Initializing Database Connection...")
	database.InitDB("data.db")

	// 1. Initialize TCP Sync Server (Port 9090)
	tcpServer := tcp.NewTCPServer("9090")
	go tcpServer.Start()

	// 2. Initialize UDP Notification Broadcaster (Port 9091)
	udpServer := &udp.NotificationServer{
		Port:    "9091",
		Clients: make([]net.UDPAddr, 0),
	}
	// We run this in a goroutine so it can receive 'subscribe' packets in background
	go udpServer.Start()

	// 3. Initialize WebSocket Chat Hub
	chatHub := websocket.NewHub()
	go chatHub.Run()

	server := &APIServer{
		Router:    gin.Default(),
		Database:  database.DB,
		JWTSecret: "super-secret-manga-key-for-academic-purposes-only",
	}

	// --- Public Routes ---
	authGroup := server.Router.Group("/auth")
	{
		authGroup.POST("/register", auth.RegisterUser(server.Database))
		authGroup.POST("/login", auth.LoginUser(server.Database, server.JWTSecret))
	}

	server.Router.GET("/manga", manga.SearchManga(server.Database))
	server.Router.GET("/manga/:id", manga.GetManga(server.Database))
	
	// WebSocket Endpoint
	server.Router.GET("/chat", websocket.ServeWs(chatHub))

	// --- Protected User Routes (Requires JWT) ---
	protected := server.Router.Group("/users")
	protected.Use(auth.AuthMiddleware(server.JWTSecret))
	{
		protected.POST("/library", manga.AddToLibrary(server.Database))
		
		// HTTP -> TCP Bridge: Updates DB and broadcasts to TCP clients
		protected.PUT("/progress", manga.UpdateProgress(server.Database, tcpServer))
		
		// HTTP -> UDP Bridge: Triggers a global UDP notification
		// THIS FIXES YOUR 404 ERROR
		protected.POST("/notify", manga.TriggerNotification(udpServer)) 
	}

	log.Println("🚀 MangaHub API Server starting on port 8080...")
	if err := server.Router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}