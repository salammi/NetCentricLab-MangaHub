package manga

import (
	"NetCentricLab-MangaHub/internal/tcp"
	"NetCentricLab-MangaHub/internal/udp"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchManga(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(200, gin.H{"message": "Search endpoint"}) }
}

func GetManga(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(200, gin.H{"message": "Get Manga endpoint"}) }
}

func AddToLibrary(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(200, gin.H{"message": "Added to library"}) }
}

func UpdateProgress(db *sql.DB, tcpServer *tcp.ProgressSyncServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MangaID string `json:"manga_id"`
			Chapter int    `json:"chapter"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
			return
		}
		// Trigger the TCP Broadcast!
		tcpServer.Broadcast(req)
		c.JSON(http.StatusOK, gin.H{"message": "Progress synchronized successfully!"})
	}
}
func TriggerNotification(udpServer *udp.NotificationServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MangaID string `json:"manga_id"`
			Message string `json:"message"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification data"})
			return
		}

		// Use the UDP server instance to broadcast the message
		udpServer.Broadcast(req.MangaID, req.Message)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Global UDP notification broadcasted",
		})
	}
}