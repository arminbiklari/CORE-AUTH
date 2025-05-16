package api

import (
	"core-auth/handlers/auth"
	"core-auth/handlers/health"
	"core-auth/handlers/user"
	"core-auth/internal/oauth2"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	// Initialize handlers
	userHandler := user.NewUserHandler(db)
	healthHandler := health.NewHealthHandler(db, rdb)
	authHandler := auth.NewAuthHandler(db)
	oauth2Handler := auth.NewOAuth2ServerHandler(oauth2.NewServer(rdb, db), oauth2.NewManager(rdb, db), db, rdb)
	// --- Health check ---
	router.GET("/health", healthHandler.Check)
	// --- Traditional Auth (Login, Refresh for UI/Direct Users) ---
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", userHandler.CreateUser)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		// authGroup.POST("/logout", authHandler.Logout)
	}
	// --- OAuth2 Server Endpoints ---
	oauth2Group := router.Group("/oauth2")
	{
		// Authorization endpoint (Step A)
		oauth2Group.GET("/authorize", oauth2Handler.Authorize)
		
		// Authorization callback (Step B)
		oauth2Group.GET("/callback", oauth2Handler.Authorize)
		
		// Token endpoint (Step D)
		oauth2Group.POST("/token", oauth2Handler.Token)
		
		// Token validation (Step F)
		oauth2Group.GET("/validate", oauth2Handler.Token)
	}

	// // User routes
	// users := router.Group("/users")
	// {
	// 	users.GET("/profile", userHandler.GetUser)
	// 	users.PUT("/profile", userHandler.UpdateUser)
	// }

	// // Protected routes (using traditional auth)
	// protected := router.Group("/api")
	// protected.Use(auth.AuthMiddleware())
	// {
	// 	protected.GET("/me", userHandler.GetCurrentUser)
	// }
}
