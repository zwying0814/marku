package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type likeRequest struct {
	Path      string `json:"path"`
	VisitorID string `json:"visitorId"`
}

type likeQuery struct {
	Path      string `form:"path" binding:"required"`
	VisitorID string `form:"visitorId"`
}

func TogglePublicLike(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		request, ok := bindLikeRequest(c)
		if !ok {
			return
		}
		state, err := services.ToggleLike(site, request.Path, request.VisitorID)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, state)
	}
}

func SetPublicLike(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		request, ok := bindLikeRequest(c)
		if !ok {
			return
		}
		state, err := services.SetLike(site, request.Path, request.VisitorID, true)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, state)
	}
}

func UnsetPublicLike(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		request, ok := bindLikeRequest(c)
		if !ok {
			return
		}
		state, err := services.SetLike(site, request.Path, request.VisitorID, false)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, state)
	}
}

func GetPublicLike(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		var query likeQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			writeBadRequest(c, "path is required")
			return
		}
		state, err := services.GetLikeState(site.ID, query.Path, query.VisitorID)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, state)
	}
}

func bindLikeRequest(c *gin.Context) (likeRequest, bool) {
	var request likeRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return likeRequest{}, false
		}
	}
	if request.Path == "" {
		request.Path = c.Query("path")
	}
	if request.VisitorID == "" {
		request.VisitorID = c.Query("visitorId")
	}
	return request, true
}
