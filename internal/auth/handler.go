// internal/auth/handler.go
package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"NetCentricLab-MangaHub/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Register handles POST /auth/register
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or weak password. Must be at least 8 characters."})
			return
		}

		// Hash password using bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Generate a simple unique user ID
		userID := fmt.Sprintf("usr_%d", time.Now().Unix())

		// Insert into SQLite database
		_, err = db.Exec("INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)", userID, req.Username, string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Username '%s' already exists", req.Username)})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Account created successfully!",
			"user_id":  userID,
			"username": req.Username,
			"email":    req.Email, // Email is accepted but not stored in the core schema
		})
	}
}

// Login handles POST /auth/login
func Login(db *sql.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		var storedHash, userID string
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", req.Username).Scan(&userID, &storedHash)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found"})
			return
		}

		// Compare passwords
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate JWT Token valid for 24 hours
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  userID,
			"username": req.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, _ := token.SignedString([]byte(jwtSecret))

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful!",
			"token":   tokenString,
		})
	}
}
