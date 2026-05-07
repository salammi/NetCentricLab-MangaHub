// pkg/models/manga.go
package models

// Manga represents the core manga details matching the Simplified Data Structure
type Manga struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Genres        string `json:"genres"` // Stored as a JSON array string in SQLite
	Status        string `json:"status"`
	TotalChapters int    `json:"total_chapters"`
	Description   string `json:"description"`
	CoverURL      string `json:"cover_url"` // Newly added field
}

// LibraryAddRequest maps to the CLI: mangahub library add --manga-id <id> --status <status>
type LibraryAddRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Status  string `json:"status" binding:"required"`
}
