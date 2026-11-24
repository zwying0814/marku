package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"marku/api/internal/model"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type createSiteRequest struct {
	Name              string   `json:"name" binding:"required"`
	Slug              string   `json:"slug" binding:"required"`
	Domain            string   `json:"domain"`
	AllowedOrigins    []string `json:"allowedOrigins"`
	CommentModeration string   `json:"commentModeration"`
}

type updateSiteRequest struct {
	Name              *string   `json:"name"`
	Domain            *string   `json:"domain"`
	Status            *string   `json:"status"`
	AllowedOrigins    *[]string `json:"allowedOrigins"`
	CommentModeration *string   `json:"commentModeration"`
}

type listSitesQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Q        string `form:"q"`
}

type siteResponse struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	Slug              string    `json:"slug"`
	Domain            string    `json:"domain"`
	Status            string    `json:"status"`
	SiteKeyPrefix     string    `json:"siteKeyPrefix"`
	AllowedOrigins    []string  `json:"allowedOrigins"`
	CommentModeration string    `json:"commentModeration"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	SiteKey           string    `json:"siteKey,omitempty"`
}

func CreateSite(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createSiteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		site, key, err := services.CreateSite(service.CreateSiteInput{
			Name:              request.Name,
			Slug:              request.Slug,
			Domain:            request.Domain,
			AllowedOrigins:    request.AllowedOrigins,
			CommentModeration: request.CommentModeration,
		})
		if err != nil {
			writeServiceError(c, err)
			return
		}

		result := newSiteResponse(site)
		result.SiteKey = key
		response.Created(c, result)
	}
}

func ListSites(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query listSitesQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			writeBadRequest(c, "invalid query parameters")
			return
		}
		if query.Page < 1 {
			query.Page = 1
		}
		if query.PageSize < 1 {
			query.PageSize = 20
		}
		if query.PageSize > 100 {
			query.PageSize = 100
		}

		sites, total, err := services.ListSites(query.Page, query.PageSize, query.Q)
		if err != nil {
			writeServiceError(c, err)
			return
		}

		items := make([]siteResponse, 0, len(sites))
		for index := range sites {
			items = append(items, newSiteResponse(&sites[index]))
		}

		response.OK(c, gin.H{
			"items":    items,
			"total":    total,
			"page":     query.Page,
			"pageSize": query.PageSize,
		})
	}
}

func GetSite(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseSiteID(c)
		if !ok {
			return
		}

		site, err := services.GetSite(id)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, newSiteResponse(site))
	}
}

func UpdateSite(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseSiteID(c)
		if !ok {
			return
		}

		var request updateSiteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		site, err := services.UpdateSite(id, service.UpdateSiteInput{
			Name:              request.Name,
			Domain:            request.Domain,
			Status:            request.Status,
			AllowedOrigins:    request.AllowedOrigins,
			CommentModeration: request.CommentModeration,
		})
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, newSiteResponse(site))
	}
}

func DeleteSite(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseSiteID(c)
		if !ok {
			return
		}
		if err := services.DeleteSite(id); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func RotateSiteKey(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseSiteID(c)
		if !ok {
			return
		}

		site, key, err := services.RotateSiteKey(id)
		if err != nil {
			writeServiceError(c, err)
			return
		}

		result := newSiteResponse(site)
		result.SiteKey = key
		response.OK(c, result)
	}
}

func RevokeSiteKey(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseSiteID(c)
		if !ok {
			return
		}
		if err := services.RevokeSiteKey(id); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func newSiteResponse(site *model.Site) siteResponse {
	return siteResponse{
		ID:                site.ID,
		Name:              site.Name,
		Slug:              site.Slug,
		Domain:            site.Domain,
		Status:            site.Status,
		SiteKeyPrefix:     site.SiteKeyPrefix,
		AllowedOrigins:    site.AllowedOriginList(),
		CommentModeration: site.CommentModeration,
		CreatedAt:         site.CreatedAt,
		UpdatedAt:         site.UpdatedAt,
	}
}

func parseSiteID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		writeBadRequest(c, "invalid site id")
		return 0, false
	}
	return uint(value), true
}
