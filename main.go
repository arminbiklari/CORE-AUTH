package main

import (
	"core-auth/api"
	cache "core-auth/internal/cache"
	database "core-auth/db"
	"log"
)

func main() {
	rdb, err := cache.InitRedis()
	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	if err := api.InitRoutes(db, rdb); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
