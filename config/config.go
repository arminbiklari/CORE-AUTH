package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
		GinMode string `json:"gin_mode"`
	} `json:"server"`
	
	Database struct {
		Host         string `json:"host"`
		Port         string `json:"port"`
		Name         string `json:"name"`
		User         string `json:"user"`
		Password     string `json:"password"`
		MaxIdleConns int    `json:"max_idle_conns"`
		MaxOpenConns int    `json:"max_open_conns"`
		ConnTimeout      int `json:"conn_timeout"`       // in seconds
		ReadTimeout      int `json:"read_timeout"`       // in seconds
		WriteTimeout     int `json:"write_timeout"`      // in seconds
		ConnMaxLifetime  int `json:"conn_max_lifetime"`  // in minutes
		ConnMaxIdleTime  int `json:"conn_max_idle_time"` // in minutes
	} `json:"database"`
	
	JWT struct {
		Secret       string `json:"secret"`
		ExpiryHours  int    `json:"expiry_hours"`
		RefreshHours int    `json:"refresh_hours"`
	} `json:"jwt"`
}

// Load loads configuration from a JSON file
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{}
	
	// Server config
	config.Server.Port = getEnvOrDefault("SERVER_PORT", "8080")
	config.Server.Host = getEnvOrDefault("SERVER_HOST", "0.0.0.0")
	config.Server.GinMode = getEnvOrDefault("GIN_MODE", "release")
	
	// Database config
	config.Database.Host = getEnvOrDefault("DB_HOST", "localhost")
	config.Database.Port = getEnvOrDefault("DB_PORT", "3306")
	config.Database.Name = getEnvOrDefault("DB_NAME", "auth_db")
	config.Database.User = getEnvOrDefault("DB_USER", "root")
	config.Database.Password = getEnvOrDefault("DB_PASSWORD", "")
	config.Database.MaxIdleConns = getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 10)
	config.Database.MaxOpenConns = getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 100)
	config.Database.ConnTimeout = getEnvAsIntOrDefault("DB_CONN_TIMEOUT", 5)
	config.Database.ReadTimeout = getEnvAsIntOrDefault("DB_READ_TIMEOUT", 30)
	config.Database.WriteTimeout = getEnvAsIntOrDefault("DB_WRITE_TIMEOUT", 30)
	config.Database.ConnMaxLifetime = getEnvAsIntOrDefault("DB_CONN_MAX_LIFETIME", 60)
	config.Database.ConnMaxIdleTime = getEnvAsIntOrDefault("DB_CONN_MAX_IDLE_TIME", 5)
	
	// JWT config
	config.JWT.Secret = getEnvOrDefault("JWT_SECRET", "your-secret-key")
	config.JWT.ExpiryHours = getEnvAsIntOrDefault("JWT_EXPIRY_HOURS", 24)
	config.JWT.RefreshHours = getEnvAsIntOrDefault("JWT_REFRESH_HOURS", 168)

	return config, nil
}