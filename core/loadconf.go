package core

import (
	"core-auth/config"
)

func DbConfig() *config.Config {
	dbConfig, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
	}
	return dbConfig
}

func ServerConfig() *config.Config {
	serverConfig, err := config.LoadFromEnv()
	if err != nil {
		panic(err)
	}
	return serverConfig
}
