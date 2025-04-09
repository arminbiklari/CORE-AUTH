package auth

import (
	database "core-auth/db"
	token "core-auth/internal/tokens"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // in seconds
}

// Login generates only refresh token
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by username
	if !database.CheckUsernameDB(h.db, req.Username) {
		log.Printf("Login failed: username %s not found", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !database.CheckPasswordDB(h.db, req.Username, req.Password) {
		log.Printf("Login failed: invalid password for user %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	user, err := database.GetUserByUsername(h.db, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// Check password is correct CheckPasswordDB
	if !database.CheckPasswordDB(h.db, req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if !user.CheckActive(h.db) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is not active"})
		return
	}
	// Generate refresh token
	refreshToken, tokenExpiry, err := token.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Store refresh token in database
	if err := database.StoreRefreshToken(h.db, req.Username, refreshToken, tokenExpiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		RefreshToken: refreshToken,
		ExpiresIn:   int(time.Until(tokenExpiry).Seconds()),
	})
}