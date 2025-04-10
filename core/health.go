package core

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type HealthStatus struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

func CheckHealth(db *gorm.DB) (*HealthStatus, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Database:  "error",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Database:  "error",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}, err
	}

	return &HealthStatus{
		Status:    "healthy",
		Database:  "connected",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}