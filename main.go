package main

import (
	"core-auth/api"
	database "core-auth/db"
	"log"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	if err := api.InitRoutes(db); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
