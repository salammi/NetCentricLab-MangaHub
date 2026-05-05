// pkg/models/manga.go
package models

// Manga represents the core manga details
type Manga struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Genres        string `json:"genres"` // Stored as a JSON array string in SQLite
	Status        string `json:"status"`
	TotalChapters int    `json:"total_chapters"`
	Description   string `json:"description"`
}

// LibraryAddRequest maps to the CLI: mangahub library add --manga-id <id> --status <status>
type LibraryAddRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Status  string `json:"status" binding:"required"`
}
