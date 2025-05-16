package oauth2

import (
	"core-auth/config"
	"log"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	oredis "github.com/go-oauth2/redis/v4"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Server wraps the oauth2 server with our configuration
type Server struct {
	*server.Server
}

// NewServer creates a new OAuth2 server with Redis storage
func NewServer(rdb *redis.Client, db *gorm.DB) *Server {
	config, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// Create manager
	manager := manage.NewDefaultManager()

	// Use Redis token store
	manager.MapTokenStorage(oredis.NewRedisStore(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
		DialTimeout: 10 * time.Second,
		ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}))

	// Set token configuration
	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Duration(config.OAuth2Server.AccessTokenDuration) * time.Minute,
		RefreshTokenExp:   time.Duration(config.OAuth2Server.RefreshTokenDuration) * time.Hour,
		IsGenerateRefresh: true,
	})

	// Set token generator
	manager.MapAccessGenerate(generates.NewAccessGenerate())

	// Create server
	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	// Set error handlers
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
	})

	return &Server{srv}
} 