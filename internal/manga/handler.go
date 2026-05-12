package manga

import (
	"NetCentricLab-MangaHub/internal/tcp"
	"NetCentricLab-MangaHub/internal/udp"
	"database/sql"
	"fmt"
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

		// 1. DATABASE VALIDATION: Fetch total chapters for this specific manga
		var totalChapters int
		query := "SELECT total_chapters FROM manga WHERE id = ?"
		err := db.QueryRow(query, req.MangaID).Scan(&totalChapters)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Manga ID not found in database"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}

		// 2. LOGIC CHECK: Prevent progress beyond the total
		if req.Chapter > totalChapters {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid chapter: %d exceeds manga total of %d", req.Chapter, totalChapters),
			})
			return
		}

		// 3. (Optional but recommended) Update the user_progress table in DB here
		// _, err = db.Exec("INSERT OR REPLACE INTO user_progress...")

		// 4. PROTOCOL SYNC: Only trigger the TCP Broadcast if data is valid
		tcpServer.Broadcast(req)

		c.JSON(http.StatusOK, gin.H{
			"message": "Progress synchronized successfully!",
			"details": fmt.Sprintf("Updated %s to chapter %d", req.MangaID, req.Chapter),
		})
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