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

		// 1. Get the Total Count
		var totalManga int
		var countErr error
		if query == "" {
			countErr = db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&totalManga)
		} else {
			// UPDATED: Now searches both the title and the id columns
			countErr = db.QueryRow("SELECT COUNT(*) FROM manga WHERE title LIKE ? OR id LIKE ?", "%"+query+"%", "%"+query+"%").Scan(&totalManga)
		}

		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error counting manga"})
			return
		}

		// 2. Calculate Total Pages dynamically
		totalPages := (totalManga + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}

		// 3. Fetch the actual paginated results
		var rows *sql.Rows
		var dbErr error
		if query == "" {
			rows, dbErr = db.Query("SELECT id, title, author, genres, status, total_chapters, description, cover_url FROM manga LIMIT ? OFFSET ?", limit, offset)
		} else {
			// UPDATED: Now searches both the title and the id columns.
			// Note the parentheses around the OR statement so the LIMIT applies correctly!
			rows, dbErr = db.Query("SELECT id, title, author, genres, status, total_chapters, description, cover_url FROM manga WHERE (title LIKE ? OR id LIKE ?) LIMIT ? OFFSET ?", "%"+query+"%", "%"+query+"%", limit, offset)
		}

		if dbErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching manga"})
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

		// 4. Update the JSON response using a struct to enforce exact chronological key order
		response := struct {
			Limit      int            `json:"limit"`
			TotalManga int            `json:"total_manga"`
			Page       int            `json:"page"`
			TotalPages int            `json:"total_pages"`
			Results    []models.Manga `json:"results"`
		}{
			Limit:      limit,
			TotalManga: totalManga,
			Page:       page,
			TotalPages: totalPages,
			Results:    results,
		}

		c.JSON(http.StatusOK, response)
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

// GetLibrary handles GET /users/library to fetch reading progress
func GetLibrary(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}
		userID := userIDRaw.(string)

		// Query to join progress with manga details
		rows, err := db.Query(`
			SELECT m.id, m.title, up.current_chapter, up.status, up.updated_at
			FROM user_progress up
			JOIN manga m ON up.manga_id = m.id
			WHERE up.user_id = ?`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching library"})
			return
		}
		defer rows.Close()

		var library []map[string]interface{}
		for rows.Next() {
			var id, title, status, updatedAt string
			var chapter int
			if err := rows.Scan(&id, &title, &chapter, &status, &updatedAt); err != nil {
				continue
			}
			library = append(library, map[string]interface{}{
				"manga_id":        id,
				"title":           title,
				"current_chapter": chapter,
				"status":          status,
				"last_updated":    updatedAt,
			})
		}

		if library == nil {
			library = []map[string]interface{}{}
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"library": library,
		})
	}
}
