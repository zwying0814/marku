package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error(
					"panic recovered",
					"request_id", GetRequestID(c),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "internal server error"))
			}
		}()

		c.Next()
	}
}
