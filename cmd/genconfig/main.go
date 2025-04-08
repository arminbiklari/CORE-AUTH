package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"core-auth/config"
)

func main() {
	outputPath := flag.String("output", "config.json", "Path to output configuration file")
	envFile := flag.Bool("env", false, "Generate .env file instead of JSON")
	flag.Parse()

	// Create default configuration
	cfg := &config.Config{}
	
	// Set default values
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	
	cfg.Database.Host = "localhost"
	cfg.Database.Port = "3306"
	cfg.Database.Name = "auth_db"
	cfg.Database.User = "root"
	cfg.Database.Password = ""
	cfg.Database.MaxIdleConns = 10
	cfg.Database.MaxOpenConns = 100
	
	cfg.JWT.Secret = "your-secret-key-change-this-in-production"
	cfg.JWT.ExpiryHours = 24
	cfg.JWT.RefreshHours = 168

	if *envFile {
		// Generate .env file
		envContent := fmt.Sprintf(`# Server Configuration
SERVER_PORT=%s
SERVER_HOST=%s

# Database Configuration
DB_HOST=%s
DB_PORT=%s
DB_NAME=%s
DB_USER=%s
DB_PASSWORD=%s
DB_MAX_IDLE_CONNS=%d
DB_MAX_OPEN_CONNS=%d

# JWT Configuration
JWT_SECRET=%s
`,
			cfg.Server.Port,
			cfg.Server.Host,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.MaxIdleConns,
			cfg.Database.MaxOpenConns,
			cfg.JWT.Secret,
		)

		if err := os.WriteFile(*outputPath, []byte(envContent), 0644); err != nil {
			fmt.Printf("Error writing .env file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Create output directory if it doesn't exist
		dir := filepath.Dir(*outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}

		// Generate JSON configuration
		file, err := os.Create(*outputPath)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(cfg); err != nil {
			fmt.Printf("Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Configuration file generated at: %s\n", *outputPath)
} 