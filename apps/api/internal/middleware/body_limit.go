package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
)

func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 {
			c.Next()
			return
		}

		if c.Request.ContentLength > maxBytes {
			response.Abort(c, response.New(http.StatusRequestEntityTooLarge, response.CodePayloadTooLarge, "request body too large"))
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}

		c.Next()
	}
}
