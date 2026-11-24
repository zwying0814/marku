package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"marku/api/internal/config"
	"marku/api/internal/response"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
			return
		}

		c.Writer.Header().Add("Vary", "Origin")
		_, originAllowed := allowed[origin]
		_, wildcardAllowed := allowed["*"]
		path := c.Request.URL.Path
		publicPath := path == "/api/v1/public" || strings.HasPrefix(path, "/api/v1/public/")
		if !originAllowed && !wildcardAllowed && !publicPath {
			if c.Request.Method == http.MethodOptions {
				response.Abort(c, response.New(http.StatusForbidden, response.CodeForbidden, "origin not allowed"))
				return
			}
			c.Next()
			return
		}

		if publicPath {
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
			}
		} else if wildcardAllowed && !cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After")

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Site-Key, X-Visitor-Id, X-Request-ID, Idempotency-Key")
			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(int(cfg.MaxAge.Seconds())))
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
