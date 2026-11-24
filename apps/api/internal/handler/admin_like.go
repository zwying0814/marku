package handler

import (
	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
	"marku/api/internal/service"
)

func GetSiteLikeStats(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID, ok := parseSiteID(c)
		if !ok {
			return
		}

		days, err := parseRangeDays(c.Query("range"))
		if err != nil {
			writeBadRequest(c, err.Error())
			return
		}
		limit, err := parseLimit(c.Query("limit"))
		if err != nil {
			writeBadRequest(c, err.Error())
			return
		}

		stats, err := services.GetSiteLikeStats(siteID, c.Query("path"), days, limit)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, stats)
	}
}
