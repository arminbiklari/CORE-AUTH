package user

import (
	database "core-auth/db"
	token "core-auth/internal/tokens"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RoleID   *uint  `json:"role_id"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint   `json:"role_id"`
	RefreshToken string `json:"refresh_token"`
	RefreshTokenExpiry *time.Time `json:"refresh_token_expiry"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default role if none provided
	roleID := database.GetDefaultUserRole()
	if req.RoleID != nil {
		// Validate role ID if provided
		if err := database.ValidateRoleID(h.db, *req.RoleID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		roleID = *req.RoleID
	}

	refreshToken, tokenExpiry, err := token.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	// Create new user
	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   roleID,
		RefreshToken: refreshToken,
		RefreshTokenExpiry: &tokenExpiry,
	}

	// Use the database function to create user
	if err := database.CreateUser(h.db, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	response := UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RoleID:   user.RoleID,
		RefreshToken: refreshToken,
		RefreshTokenExpiry: &tokenExpiry,
	}

	c.JSON(http.StatusCreated, response)
}
