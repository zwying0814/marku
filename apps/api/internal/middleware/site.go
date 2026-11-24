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

const siteContextKey = "marku.site"

func SiteKey(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("X-Site-Key"))
		if key == "" {
			response.Abort(c, response.New(http.StatusUnauthorized, response.CodeUnauthorized, "site key required"))
			return
		}

		site, err := services.ResolveSiteByKey(key)
		switch {
		case errors.Is(err, service.ErrInvalidSiteKey):
			response.Abort(c, response.New(http.StatusUnauthorized, response.CodeUnauthorized, "invalid site key"))
			return
		case errors.Is(err, service.ErrSiteDisabled):
			response.Abort(c, response.New(http.StatusForbidden, response.CodeForbidden, "site is disabled"))
			return
		case err != nil:
			response.Abort(c, response.Wrap(err, http.StatusInternalServerError, response.CodeInternal, "internal server error"))
			return
		}

		c.Set(siteContextKey, site)
		c.Set("marku.site_id", site.ID)
		c.Next()
	}
}

func GetSite(c *gin.Context) *model.Site {
	value, ok := c.Get(siteContextKey)
	if !ok {
		return nil
	}
	site, _ := value.(*model.Site)
	return site
}

func GetSiteID(c *gin.Context) (uint, bool) {
	value, ok := c.Get("marku.site_id")
	if !ok {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}
