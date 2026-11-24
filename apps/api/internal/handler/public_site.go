package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/response"
)

type publicSiteResponse struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	Domain            string `json:"domain"`
	CommentModeration string `json:"commentModeration"`
}

func PublicSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		response.OK(c, publicSiteResponse{
			ID:                site.ID,
			Name:              site.Name,
			Slug:              site.Slug,
			Domain:            site.Domain,
			CommentModeration: site.CommentModeration,
		})
	}
}
