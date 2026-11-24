package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
)

func SiteOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		site := GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		origins := site.AllowedOriginList()
		if len(origins) == 0 {
			c.Next()
			return
		}

		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}
		for _, allowed := range origins {
			if origin == allowed {
				c.Next()
				return
			}
		}

		response.Abort(c, response.New(http.StatusForbidden, response.CodeForbidden, "origin not allowed"))
	}
}
