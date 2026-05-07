// pkg/models/user.go
package models

// RegisterRequest maps to the CLI: mangahub auth register --username <user> --email <email>
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest maps to the CLI: mangahub auth login --username <user>
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
