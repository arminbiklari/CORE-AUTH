package api

import (
	"log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"core-auth/config"
)

func InitRoutes(db *gorm.DB) error {
	config, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	gin.SetMode(config.Server.GinMode)
	router := gin.Default()
	SetupRoutes(router, db)

	addr := config.Server.Host + ":" + config.Server.Port
	log.Printf("Starting server on %s", addr)
	return router.Run(addr)
}