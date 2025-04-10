package health

import (
	"core-auth/internal/core"
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	status, err := core.CheckHealth(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, status)
		return
	}
	c.JSON(http.StatusOK, status)
}