package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"marku/api/internal/database"
	"marku/api/internal/response"
	"marku/api/internal/version"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "ok",
		"service": "marku-api",
		"version": version.Current,
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
