package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

func GetPublicPageStats(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		items, err := services.GetPublicPageStats(site.ID, collectPathQuery(c), c.Query("visitorId"))
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, gin.H{"items": items})
	}
}
