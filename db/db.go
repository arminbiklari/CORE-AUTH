package database

import (
	"context"
	"core-auth/config"
	"core-auth/credentials"
	"core-auth/vault"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Interface for a credential provider
type CredentialProvider interface {
	GetCredentials() (string, string)
}

// Default credential provider using environment variables
type envCredentialProvider struct {
	cfg *config.Config
}

func (e *envCredentialProvider) GetCredentials() (string, string) {
	return e.cfg.Database.User, e.cfg.Database.Password
}

// Global credential provider that can be set from outside
var (
	credentialProvider CredentialProvider
	providerMutex      sync.RWMutex
)

// SetCredentialProvider allows setting a credential provider from outside packages
func SetCredentialProvider(provider CredentialProvider) {
	providerMutex.Lock()
	defer providerMutex.Unlock()
	credentialProvider = provider
}

// Global DB reference for connection management
var (
	globalDBMutex sync.RWMutex
	globalDB      *gorm.DB
)

// getDatabaseCredentials retrieves credentials from the credentials package or directly from config
func getDatabaseCredentials(cfg *config.Config) (string, string, error) {
	// Try to get credentials from the credentials provider first
	username, password := credentials.Get()
	
	if username != "" && password != "" {
		log.Println("Using database credentials from credentials provider")
		return username, password, nil
	}
	
	// Fall back to direct Vault access if configured
	if cfg.Vault.Token != "" && cfg.Vault.RolePath != "" {
		log.Println("Credentials provider not available, fetching directly from Vault")
		
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

		log.Println("Successfully retrieved database credentials directly from Vault")
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

	// Set the global DB instance
	replaceDBInstance(db)

	return db, nil
}

// RenewConnection updates the database connection with new credentials
func RenewConnection(username, password string) error {
	// Get the current configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Build a new DSN with the new credentials
	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		username,
		password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.ConnTimeout,
		cfg.Database.ReadTimeout,
		cfg.Database.WriteTimeout,
	)

	// Create a new connection with the new credentials
	newDB, err := gorm.Open(mysql.Open(dbDsn), &gorm.Config{
		PrepareStmt: true,
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      logger.Warn,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		return fmt.Errorf("failed to open new database connection: %v", err)
	}

	// Test the new connection
	sqlDB, err := newDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database with new credentials: %v", err)
	}

	// Apply connection settings
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(time.Minute * time.Duration(cfg.Database.ConnMaxIdleTime))
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(cfg.Database.ConnMaxLifetime))

	// Replace the global DB instance
	// Note: This assumes you have a way to access and update the global DB instance
	// You may need to adjust this based on how your application manages DB connections
	replaceDBInstance(newDB)

	log.Printf("Successfully renewed database connection with new credentials")
	return nil
}

// GetDB returns the current database connection
func GetDB() *gorm.DB {
	globalDBMutex.RLock()
	defer globalDBMutex.RUnlock()
	return globalDB
}

// replaceDBInstance updates the global DB reference with a new connection
func replaceDBInstance(db *gorm.DB) {
	globalDBMutex.Lock()
	defer globalDBMutex.Unlock()
	globalDB = db
}
