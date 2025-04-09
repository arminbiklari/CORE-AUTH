package database

import (
	"core-auth/config"
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	// get database config from config.go file 
	dbConfig, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
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
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}
	log.Printf("Successfully connected to database at %s:%s", dbConfig.Database.Host, dbConfig.Database.Port)
	sqlDB.SetMaxIdleConns(dbConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.Database.MaxOpenConns)
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
