package handler

import (
	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
	"marku/api/internal/service"
)

func GetAdminOverview(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, err := parseRangeDays(c.Query("range"))
		if err != nil {
			writeBadRequest(c, err.Error())
			return
		}
		overview, err := services.GetAdminOverview(days)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, overview)
	}
}
