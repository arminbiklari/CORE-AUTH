package database

import (
	"core-auth/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	// get database config from config.go file 
	dbConfig, err := config.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
		dbConfig.Database.User, 
		dbConfig.Database.Password, 
		dbConfig.Database.Host, 
		dbConfig.Database.Port, 
		dbConfig.Database.Name,
	)
	
	db, err := gorm.Open(mysql.Open(dbDsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(dbConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.Database.MaxOpenConns)

	// Auto-migrate the schema
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %v", err)
	}

	// Initialize default roles
	if err := InitializeRoles(db); err != nil {
		return nil, fmt.Errorf("failed to initialize roles: %v", err)
	}

	return db, nil
}
