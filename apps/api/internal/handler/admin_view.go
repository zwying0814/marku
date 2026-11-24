package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
	"marku/api/internal/service"
)

func GetSiteViewStats(services *service.Service) gin.HandlerFunc {
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

		stats, err := services.GetSiteViewStats(siteID, days, limit)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, stats)
	}
}

func parseRangeDays(raw string) (int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return 7, nil
	}
	value = strings.TrimSuffix(value, "d")
	days, err := strconv.Atoi(value)
	if err != nil || days < 1 || days > 365 {
		return 0, fmt.Errorf("range must be between 1d and 365d")
	}
	return days, nil
}

func parseLimit(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 10, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > 50 {
		return 0, fmt.Errorf("limit must be between 1 and 50")
	}
	return limit, nil
}
