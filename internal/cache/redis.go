package cache

import (
	"context"
	"core-auth/config"
	"fmt"
	"log"
	"time"
	"github.com/redis/go-redis/v9"
)

// InitRedis initializes and returns a Redis client instance.
func InitRedis() (*redis.Client, error) {
	conf, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load redis config: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})

	// Ping the Redis server to ensure connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Successfully connected to Redis at %s", conf.Redis.Addr)
	return rdb, nil
}

// TODO: add helper functions here later for common Redis operations
// like setting keys with expiration, getting keys, etc. 