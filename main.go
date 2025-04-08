package main

import (
	"encoding/json"
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

	// Debug: Print loaded configuration
	if configJSON, err := json.MarshalIndent(conf, "", "  "); err == nil {
		log.Printf("Loaded configuration:\n%s", string(configJSON))
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	// Print database connection status
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err) 
	}
	log.Printf("Successfully connected to database at %s:%s", conf.Database.Host, conf.Database.Port)

	// Create Schema and migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
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
