package main

import (
	"log"
	"core-auth/api"
	"core-auth/config"
	database "core-auth/db"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Load configuration
	conf, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		panic(err)
	}

	// Create router
	router := gin.Default()

	// Set Gin mode from configuration
	gin.SetMode(conf.Server.GinMode)

	// Setup routes
	api.SetupRoutes(router, db)

	// Start the server
	port := conf.Server.Port
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
