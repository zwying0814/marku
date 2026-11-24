package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type recordViewRequest struct {
	Path      string `json:"path" binding:"required"`
	VisitorID string `json:"visitorId"`
}

func RecordPublicView(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		var request recordViewRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		result, err := services.RecordPageView(site, request.Path, request.VisitorID, c.Request.UserAgent())
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, result)
	}
}

func GetPublicViews(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		paths := collectPathQuery(c)

		items, err := services.GetPageViews(site.ID, paths)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, gin.H{"items": items})
	}
}

func collectPathQuery(c *gin.Context) []string {
	values := c.QueryArray("paths")
	if len(values) == 0 {
		if path := c.Query("path"); path != "" {
			values = []string{path}
		}
	}

	paths := make([]string, 0, len(values))
	for _, value := range values {
		for _, path := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(path); trimmed != "" {
				paths = append(paths, trimmed)
			}
		}
	}
	return paths
}
