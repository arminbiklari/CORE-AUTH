package api

import (
	auth "core-auth/handlers/auth"
	user "core-auth/handlers/user"
	health "core-auth/handlers/health"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	userHandler := user.NewUserHandler(db)
	authHandler := auth.NewAuthHandler(db)
	healthHandler := health.NewHealthHandler(db)

	// Health check endpoint
	router.GET("/health", healthHandler.Check)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no middleware)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login) // login with username and password and get refresh token
			auth.POST("/refresh", authHandler.RefreshToken) // refresh token and get access token
		}
		// Protected routes
		users := v1.Group("/users/register")
		{
			users.POST("", userHandler.CreateUser)
		}
	}
}
