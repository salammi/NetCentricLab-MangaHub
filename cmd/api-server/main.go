// cmd/api-server/main.go
package main

import (
	"database/sql"
	"log"

	"NetCentricLab-MangaHub/internal/auth"
	"NetCentricLab-MangaHub/internal/manga"
	"NetCentricLab-MangaHub/internal/tcp"
	"NetCentricLab-MangaHub/internal/udp"
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

	// Initialize and start network servers concurrently
	tcpServer := tcp.NewServer("9090")
	go tcpServer.Start()

	udpServer := udp.NewServer("9091")
	go udpServer.Start()

	server := &APIServer{
		Router:    gin.Default(),
		Database:  database.DB,
		JWTSecret: "super-secret-manga-key-for-academic-purposes-only",
	}

	authGroup := server.Router.Group("/auth")
	{
		authGroup.POST("/register", auth.Register(server.Database))
		authGroup.POST("/login", auth.Login(server.Database, server.JWTSecret))
	}

	server.Router.GET("/manga", manga.SearchManga(server.Database))
	server.Router.GET("/manga/:id", manga.GetManga(server.Database))

	protected := server.Router.Group("/users")
	protected.Use(auth.AuthMiddleware(server.JWTSecret))
	{
		protected.POST("/library", manga.AddToLibrary(server.Database))
		protected.PUT("/progress", manga.UpdateProgress(server.Database, tcpServer))
		// New Integrated Endpoint for UDP Notifications
		protected.POST("/notify", manga.TriggerNotification(udpServer))
	}

	log.Println("Starting HTTP API Server on port 8080...")
	if err := server.Router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
