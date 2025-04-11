package auth

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/server"
	"core-auth/internal/oauth2"
)

type OAuth2ServerHandler struct {
	server  *server.Server
	manager *oauth2.Manager
	db      *gorm.DB
	rdb     *redis.Client
}

func NewOAuth2ServerHandler(server *server.Server, manager *oauth2.Manager, db *gorm.DB, rdb *redis.Client) *OAuth2ServerHandler {
	return &OAuth2ServerHandler{
		server:  server,
		manager: manager,
		db:      db,
		rdb:     rdb,
	}
}

func (h *OAuth2ServerHandler) Authorize(c *gin.Context) {
	
	
}
