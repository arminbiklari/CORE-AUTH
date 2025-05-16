package cache

import (
	"context"
	"core-auth/config"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	redisError  error
)

// InitRedis initializes and returns a Redis client instance.
func InitRedis() (*redis.Client, error) {
	if redisClient != nil {
		return redisClient, redisError
	}

	conf, err := config.LoadFromEnv()
	if err != nil {
		log.Printf("Failed to load redis config: %v", err)
		return nil, err
	}

	// Create Redis client with authentication
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
		// Add timeouts to prevent hanging
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		// Add retry mechanism
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 1 * time.Second,
	})

	// Try to connect with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		// Log the error but don't fail the application
		log.Printf("Warning: Redis connection failed: %v", err)
		redisError = err
	} else {
		log.Printf("Successfully connected to Redis at %s", conf.Redis.Addr)
		redisError = nil
	}

	redisClient = rdb
	return rdb, redisError
}

// IsRedisAvailable checks if Redis is available
func IsRedisAvailable() bool {
	if redisClient == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := redisClient.Ping(ctx).Result()
	return err == nil
}

// TODO: add helper functions here later for common Redis operations
// like setting keys with expiration, getting keys, etc. 