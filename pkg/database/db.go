// pkg/database/db.go
package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB

// InitDB opens a connection to the SQLite database and verifies it
func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection is alive
	if err = DB.Ping(); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	log.Println("Successfully connected to SQLite database")
}
