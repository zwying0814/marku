package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"marku/api/internal/database"
	"marku/api/internal/response"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "ok",
		"service": "marku-api",
		"version": "0.1.0",
	})
}

func Ready(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := database.Ping(db); err != nil {
			response.Abort(c, response.Wrap(err, http.StatusServiceUnavailable, response.CodeUnavailable, "database not ready"))
			return
		}
		response.OK(c, gin.H{"status": "ready"})
	}
}
