package main

import (
	"core-auth/api"
	"core-auth/config"
	"core-auth/credentials"
	database "core-auth/db"
	cache "core-auth/internal/cache"
	"log"
	"os"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize credentials provider
	if cfg.Vault.Token != "" && cfg.Vault.RolePath != "" {
		log.Println("Initializing Vault credentials provider")
		vaultProvider := credentials.NewVaultProvider(cfg)
		if err := vaultProvider.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize Vault credentials: %v", err)
			// Fall back to environment credentials
			credentials.RegisterEnvironmentProvider(cfg)
		} else {
			// Start credential rotation in background
			go vaultProvider.StartRotation()
		}
	} else {
		// Use environment credentials
		credentials.RegisterEnvironmentProvider(cfg)
	}

	// Initialize Redis (will continue even if Redis is unavailable)
	rdb, err := cache.InitRedis()
	if err != nil {
		log.Printf("Warning: Redis initialization failed: %v", err)
		log.Println("Application will continue in degraded mode (database-only)")
	}

	// Initialize database (required)
	db, err := database.InitDB()
	if err != nil {
		log.Printf("Error: Database initialization failed: %v", err)
		log.Println("Please check your database configuration:")
		log.Println("- Ensure DB_HOST, DB_PORT, DB_NAME, DB_USER, and DB_PASSWORD are set correctly")
		log.Println("- Verify the database server is running and accessible")
		log.Println("- Check if the user has proper permissions")
		os.Exit(1)
	}

	// Initialize routes
	if err := api.InitRoutes(db, rdb); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
