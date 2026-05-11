// internal/user/handler.go
package user

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" // Updated to v5 to match standard go.mod
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- Define the Request and Response structures locally ---
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token,omitempty"`
}
// -----------------------------------------------------------

// RegisterUser handles POST /users/register
func RegisterUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload or missing fields"})
			return
		}

		// Check if user already exists
		var existingUser string
		err := db.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", req.Email, req.Username).Scan(&existingUser)
		if err != sql.ErrNoRows {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			return
		}

		// Hash the password securely
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
			return
		}

		// Generate a unique user ID
		userID := "usr_" + uuid.New().String()

		// FIXED: Removed 'created_at' to match the database table schema
		_, err = db.Exec(`
			INSERT INTO users (id, username, email, password_hash) 
			VALUES (?, ?, ?, ?)`,
			userID, req.Username, req.Email, string(hashedPassword))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"user": UserResponse{
				ID:       userID,
				Username: req.Username,
				Email:    req.Email,
			},
		})
	}
}

// LoginUser handles POST /users/login
func LoginUser(db *sql.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		var user UserResponse
		var hashedPassword string

		// Query the database for the user by email
		err := db.QueryRow("SELECT id, username, email, password_hash FROM users WHERE email = ?", req.Email).
			Scan(&user.ID, &user.Username, &user.Email, &hashedPassword)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during login"})
			return
		}

		// Compare the provided password with the stored bcrypt hash
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
			return
		}

		user.Token = tokenString

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"user":    user,
		})
	}
}