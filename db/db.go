package database

import (
	"context"
	"core-auth/config"
	"core-auth/vault"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getDatabaseCredentials retrieves credentials either from env vars or Vault
func getDatabaseCredentials(cfg *config.Config) (string, string, error) {
	// Check if Vault is configured
	fmt.Println("Vault token:", cfg.Vault.Token)
	fmt.Println("Vault role path:", cfg.Vault.RolePath)
	if cfg.Vault.Token != "" && cfg.Vault.RolePath != "" {
		// Initialize Vault client
		vaultClient, err := vault.NewVaultClient(cfg)
		if err != nil {
			return "", "", fmt.Errorf("failed to create Vault client: %v", err)
		}

		// Get dynamic credentials from Vault
		username, password, err := vaultClient.GetDatabaseCredentials(context.Background())
		if err != nil {
			return "", "", fmt.Errorf("failed to get credentials from Vault: %v", err)
		}

		log.Println("Successfully retrieved database credentials from Vault")
		return username, password, nil
	}

	// Fall back to environment variables
	log.Println("Using database credentials from environment variables")
	return cfg.Database.User, cfg.Database.Password, nil
}

func InitDB() (*gorm.DB, error) {
	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	// Get database credentials (either from env or Vault)
	dbUser, dbPassword, err := getDatabaseCredentials(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get database credentials: %v", err)
	}

	// Validate required database configuration
	if dbUser == "" {
		return nil, fmt.Errorf("database user is required")
	}
	if dbPassword == "" {
		return nil, fmt.Errorf("database password is required")
	}
	if cfg.Database.Host == "" {
		return nil, fmt.Errorf("database host is required")
	}
	if cfg.Database.Port == "" {
		return nil, fmt.Errorf("database port is required")
	}
	if cfg.Database.Name == "" {
		return nil, fmt.Errorf("database name is required")
	}

	// Build DSN with all parameters
	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		dbUser,
		dbPassword,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.ConnTimeout,
		cfg.Database.ReadTimeout,
		cfg.Database.WriteTimeout,
	)

	// Open database connection with custom config
	db, err := gorm.Open(mysql.Open(dbDsn), &gorm.Config{
		PrepareStmt: true,
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      logger.Warn,
				IgnoreRecordNotFoundError: true, // Suppress "record not found" logs
			},
		),
	})
	if err != nil {
		panic(err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(time.Minute * time.Duration(cfg.Database.ConnMaxIdleTime))
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(cfg.Database.ConnMaxLifetime))

	// Verify connection with retry
	var pingErr error
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr = sqlDB.PingContext(ctx)
		cancel()
		if pingErr == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping database after retries: %v", pingErr)
	}

	log.Printf("Successfully connected to database at %s:%s", cfg.Database.Host, cfg.Database.Port)

	// Auto-migrate the schema
	if err := AutoMigrate(db); err != nil {
		panic(err)
	}

	// Initialize default roles
	if err := InitializeRoles(db); err != nil {
		panic(err)
	}

	return db, nil
}
