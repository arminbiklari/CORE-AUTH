package core

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type HealthStatus struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Redis     string `json:"redis"`
	Timestamp string `json:"timestamp"`
}

var (
	redisStatus     bool
	redisStatusLock sync.RWMutex
)

func init() {
	redisStatus = true // Assume Redis is up initially
}

func IsRedisUp() bool {
	redisStatusLock.RLock()
	defer redisStatusLock.RUnlock()
	return redisStatus
}

func RedisUp(rdb *redis.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	err := rdb.Ping(ctx).Err()
	
	redisStatusLock.Lock()
	redisStatus = err == nil
	redisStatusLock.Unlock()
	
	return redisStatus
}

func CheckHealth(db *gorm.DB, rdb *redis.Client) (*HealthStatus, error, bool) {
	Status := "healthy"
	
	// Check Redis connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	redisStatus := "connected"
	if err := rdb.Ping(ctx).Err(); err != nil {
		redisStatus = "error"
		Status = "unhealthy but working with database"
	}

	// Check database connection
	databaseStatus := "connected"
	sqlDB, err := db.DB()
	if err != nil {
		databaseStatus = "error"
		Status = "unhealthy"
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		databaseStatus = "error"
		Status = "unhealthy"
		log.Println("Database connection error:", err)
	}

	return &HealthStatus{
		Status:    Status,
		Database:  databaseStatus,
		Redis:     redisStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil, IsRedisUp()
}