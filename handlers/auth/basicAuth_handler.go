package auth

import (
	database "core-auth/db"
	token "core-auth/internal/tokens"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !database.CheckPasswordDB(h.db, req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
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