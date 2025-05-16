package auth

import (
	"net/http"

	"core-auth/internal/oauth2"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type OAuth2ServerHandler struct {
	server  *oauth2.Server
	manager *oauth2.Manager
	db      *gorm.DB
	rdb     *redis.Client
}

func NewOAuth2ServerHandler(server *oauth2.Server, manager *oauth2.Manager, db *gorm.DB, rdb *redis.Client) *OAuth2ServerHandler {
	return &OAuth2ServerHandler{
		server:  server,
		manager: manager,
		db:      db,
		rdb:     rdb,
	}
}

func (h *OAuth2ServerHandler) Authorize(c *gin.Context) {
	err := h.server.HandleAuthorizeRequest(c.Writer, c.Request)
	if err != nil {
		switch err {
		case errors.ErrInvalidRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		case errors.ErrUnauthorizedClient:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized_client"})
		case errors.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "access_denied"})
		case errors.ErrUnsupportedResponseType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_response_type"})
		case errors.ErrInvalidScope:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_scope"})
		case errors.ErrServerError:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		case errors.ErrTemporarilyUnavailable:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "temporarily_unavailable"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
}

func (h *OAuth2ServerHandler) Token(c *gin.Context) {
	err := h.server.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		switch err {
		case errors.ErrInvalidRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		case errors.ErrInvalidClient:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
		case errors.ErrInvalidGrant:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		case errors.ErrUnauthorizedClient:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized_client"})
		case errors.ErrUnsupportedGrantType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		case errors.ErrInvalidScope:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_scope"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
}
