package api

import (
	"log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRoutes(db *gorm.DB) error {
	router := gin.Default()
	SetupRoutes(router, db)
	log.Printf("Starting server on :8080")
	return router.Run(":8080")
}