// internal/manga/handler.go
package manga

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"NetCentricLab-MangaHub/internal/tcp"
	"NetCentricLab-MangaHub/internal/udp"
	"NetCentricLab-MangaHub/pkg/models"

	"github.com/gin-gonic/gin"
)

// SearchManga handles GET /manga?q=keyword&page=1&limit=20
func SearchManga(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 20
		}

		offset := (page - 1) * limit
		var rows *sql.Rows
		var dbErr error

		if query == "" {
			rows, dbErr = db.Query("SELECT id, title, author, genres, status, total_chapters, description, cover_url FROM manga LIMIT ? OFFSET ?", limit, offset)
		} else {
			rows, dbErr = db.Query("SELECT id, title, author, genres, status, total_chapters, description, cover_url FROM manga WHERE title LIKE ? LIMIT ? OFFSET ?", "%"+query+"%", limit, offset)
		}

		if dbErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var results []models.Manga
		for rows.Next() {
			var m models.Manga
			if err := rows.Scan(&m.ID, &m.Title, &m.Author, &m.Genres, &m.Status, &m.TotalChapters, &m.Description, &m.CoverURL); err != nil {
				continue
			}
			results = append(results, m)
		}

		if results == nil {
			results = []models.Manga{}
		}

		c.JSON(http.StatusOK, gin.H{"results": results, "page": page, "limit": limit})
	}
}

// GetManga handles GET /manga/:id
func GetManga(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var m models.Manga

		err := db.QueryRow("SELECT id, title, author, genres, status, total_chapters, description, cover_url FROM manga WHERE id = ?", id).
			Scan(&m.ID, &m.Title, &m.Author, &m.Genres, &m.Status, &m.TotalChapters, &m.Description, &m.CoverURL)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, m)
	}
}

// AddToLibrary handles POST /users/library
func AddToLibrary(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LibraryAddRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}

		_, err := db.Exec(`
			INSERT INTO user_progress (user_id, manga_id, current_chapter, status) 
			VALUES (?, ?, 0, ?)
			ON CONFLICT(user_id, manga_id) DO UPDATE SET status = excluded.status, updated_at = CURRENT_TIMESTAMP`,
			userID, req.MangaID, req.Status)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to library"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully added manga to library", "manga_id": req.MangaID, "status": req.Status})
	}
}

// ProgressUpdateRequest maps to CLI: mangahub progress update --manga-id <id> --chapter <number>
type ProgressUpdateRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Chapter int    `json:"chapter" binding:"required"`
}

// UpdateProgress handles PUT /users/progress and triggers TCP broadcasts
func UpdateProgress(db *sql.DB, tcpServer *tcp.ProgressSyncServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ProgressUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		userIDRaw, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}
		userID := userIDRaw.(string)

		// 1. Update the local SQLite database
		_, err := db.Exec(`
			INSERT INTO user_progress (user_id, manga_id, current_chapter, status, updated_at) 
			VALUES (?, ?, ?, 'reading', CURRENT_TIMESTAMP)
			ON CONFLICT(user_id, manga_id) DO UPDATE SET current_chapter = excluded.current_chapter, updated_at = CURRENT_TIMESTAMP`,
			userID, req.MangaID, req.Chapter)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during progress update"})
			return
		}

		// 2. Broadcast the update to all connected TCP clients instantly
		if tcpServer != nil {
			tcpServer.Broadcast <- tcp.ProgressUpdate{
				UserID:    userID,
				MangaID:   req.MangaID,
				Chapter:   req.Chapter,
				Timestamp: time.Now().Unix(),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Progress updated successfully and broadcasted",
			"manga_id": req.MangaID,
			"chapter":  req.Chapter,
		})
	}
}

// NotificationRequest maps to the expected payload for chapter releases
type NotificationRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// TriggerNotification handles POST /users/notify and triggers UDP broadcasts
func TriggerNotification(udpServer *udp.NotificationServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req NotificationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// UC-010: Send Chapter Release Notification
		if udpServer != nil {
			udpServer.Broadcast(udp.Notification{
				Type:      "new_chapter",
				MangaID:   req.MangaID,
				Message:   req.Message,
				Timestamp: time.Now().Unix(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Chapter release notification broadcasted successfully",
		})
	}
}
