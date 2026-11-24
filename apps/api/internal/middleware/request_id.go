package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader     = "X-Request-ID"
	requestIDContextKey = "marku.request_id"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if !requestIDPattern.MatchString(requestID) {
			requestID = newRequestID()
		}

		c.Set(requestIDContextKey, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(requestIDContextKey)
}

func newRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
