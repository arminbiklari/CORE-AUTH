package auth

import (
	database "core-auth/db"
	token "core-auth/internal/tokens"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
}

// RefreshToken validates refresh token and generates access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	if !database.CheckRefreshToken(h.db, req.RefreshToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Generate access token
	accessToken, tokenExpiry, err := token.GenerateAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:  int(time.Until(tokenExpiry).Seconds()),
	})
}
