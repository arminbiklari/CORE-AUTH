package health

import (
	"core-auth/internal/core"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:  db,
		rdb: rdb,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	status, err, redisUp := core.CheckHealth(h.db, h.rdb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, status)
		return
	}
	
	if !redisUp {
		c.JSON(http.StatusOK, gin.H{
			"status": status,
			"message": "Redis is down, using database fallback",
		})
		return
	}
	
	c.JSON(http.StatusOK, status)
}