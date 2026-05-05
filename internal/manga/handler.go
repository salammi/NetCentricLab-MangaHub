// internal/manga/handler.go
package manga

import (
	"database/sql"
	"net/http"

	"NetCentricLab-MangaHub/pkg/models"

	"github.com/gin-gonic/gin"
)

// SearchManga handles GET /manga?q=keyword
func SearchManga(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query parameter 'q' is required"})
			return
		}

		// Query SQLite using LIKE pattern for basic searching
		rows, err := db.Query("SELECT id, title, author, genres, status, total_chapters, description FROM manga WHERE title LIKE ?", "%"+query+"%")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var results []models.Manga
		for rows.Next() {
			var m models.Manga
			if err := rows.Scan(&m.ID, &m.Title, &m.Author, &m.Genres, &m.Status, &m.TotalChapters, &m.Description); err != nil {
				continue
			}
			results = append(results, m)
		}

		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}

// GetManga handles GET /manga/:id
func GetManga(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var m models.Manga

		err := db.QueryRow("SELECT id, title, author, genres, status, total_chapters, description FROM manga WHERE id = ?", id).
			Scan(&m.ID, &m.Title, &m.Author, &m.Genres, &m.Status, &m.TotalChapters, &m.Description)

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

		// Extract user_id from the JWT middleware context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			return
		}

		// Insert or replace the progress record
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
