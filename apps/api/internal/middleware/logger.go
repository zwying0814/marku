package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		status := c.Writer.Status()
		attributes := []any{
			"request_id", GetRequestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", float64(time.Since(startedAt).Microseconds()) / 1000,
			"ip", c.ClientIP(),
		}
		if route := c.FullPath(); route != "" {
			attributes = append(attributes, "route", route)
		}
		if len(c.Errors) > 0 {
			attributes = append(attributes, "error", c.Errors.String())
		}

		switch {
		case status >= 500:
			slog.Error("http request", attributes...)
		case status >= 400:
			slog.Warn("http request", attributes...)
		default:
			slog.Info("http request", attributes...)
		}
	}
}
