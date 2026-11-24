package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"marku/api/internal/model"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

const adminContextKey = "marku.admin"

func AdminAuth(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortUnauthorized(c, "authorization token required")
			return
		}

		admin, err := services.Authenticate(parts[1])
		switch {
		case errors.Is(err, service.ErrAuthNotConfigured):
			response.Abort(c, response.New(http.StatusServiceUnavailable, response.CodeUnavailable, "authentication is not configured"))
			return
		case errors.Is(err, service.ErrAdminDisabled):
			response.Abort(c, response.New(http.StatusForbidden, response.CodeForbidden, "admin account is disabled"))
			return
		case errors.Is(err, service.ErrInvalidToken):
			abortUnauthorized(c, "invalid or expired token")
			return
		case err != nil:
			response.Abort(c, response.Wrap(err, http.StatusInternalServerError, response.CodeInternal, "internal server error"))
			return
		}

		c.Set(adminContextKey, admin)
		c.Next()
	}
}

func GetAdmin(c *gin.Context) *model.Admin {
	value, ok := c.Get(adminContextKey)
	if !ok {
		return nil
	}
	admin, _ := value.(*model.Admin)
	return admin
}

func abortUnauthorized(c *gin.Context, message string) {
	c.Header("WWW-Authenticate", "Bearer")
	response.Abort(c, response.New(http.StatusUnauthorized, response.CodeUnauthorized, message))
}
