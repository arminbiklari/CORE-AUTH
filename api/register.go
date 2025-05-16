package api

import (
	"core-auth/config"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func InitRoutes(db *gorm.DB, rdb *redis.Client) error {
	config, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	gin.SetMode(config.Server.GinMode)
	router := gin.Default()
	SetupRoutes(router, db, rdb)

	addr := config.Server.Host + ":" + config.Server.Port
	log.Printf("Starting server on %s", addr)
	return router.Run(addr)
}