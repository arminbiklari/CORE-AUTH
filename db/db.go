package database

import (
	"context"
	"core-auth/config"
	"fmt"
	"log"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	// get database config from config.go file 
	dbConfig, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	// Build DSN with timeout parameters
	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
		dbConfig.Database.User, 
		dbConfig.Database.Password, 
		dbConfig.Database.Host, 
		dbConfig.Database.Port,
		dbConfig.Database.Name,
	)

	// Open database connection with default config
	db, err := gorm.Open(mysql.Open(dbDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(dbConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(time.Minute * time.Duration(dbConfig.Database.ConnMaxIdleTime))
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(dbConfig.Database.ConnMaxLifetime))

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Successfully connected to database at %s:%s", dbConfig.Database.Host, dbConfig.Database.Port)

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
