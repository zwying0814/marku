package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type createCommentRequest struct {
	Path          string `json:"path" binding:"required"`
	ParentID      *uint  `json:"parentId"`
	AuthorName    string `json:"authorName" binding:"required"`
	AuthorEmail   string `json:"authorEmail"`
	AuthorWebsite string `json:"authorWebsite"`
	Content       string `json:"content" binding:"required"`
	Honeypot      string `json:"honeypot"`
	FormStartedAt int64  `json:"formStartedAt"`
}

type publicCommentsQuery struct {
	Path     string `form:"path" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Sort     string `form:"sort"`
}

type publicCommentCountQuery struct {
	Path string `form:"path" binding:"required"`
}

type publicCommentResponse struct {
	ID            uint                    `json:"id"`
	ParentID      *uint                   `json:"parentId"`
	AuthorName    string                  `json:"authorName"`
	AuthorWebsite string                  `json:"authorWebsite"`
	Content       string                  `json:"content"`
	IsAdmin       bool                    `json:"isAdmin"`
	CreatedAt     time.Time               `json:"createdAt"`
	Replies       []publicCommentResponse `json:"replies,omitempty"`
}

func CreatePublicComment(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		var request createCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		comment, err := services.CreatePublicComment(site, service.CreatePublicCommentInput{
			Path:          request.Path,
			ParentID:      request.ParentID,
			AuthorName:    request.AuthorName,
			AuthorEmail:   request.AuthorEmail,
			AuthorWebsite: request.AuthorWebsite,
			Content:       request.Content,
			Honeypot:      request.Honeypot,
			FormStartedAt: request.FormStartedAt,
		}, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			writeServiceError(c, err)
			return
		}

		response.Created(c, gin.H{
			"id":        comment.ID,
			"status":    comment.Status,
			"createdAt": comment.CreatedAt,
		})
	}
}

func ListPublicComments(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		var query publicCommentsQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			writeBadRequest(c, "path is required")
			return
		}

		page, err := services.ListPublicComments(site.ID, query.Path, query.Page, query.PageSize, query.Sort)
		if err != nil {
			writeServiceError(c, err)
			return
		}

		items := make([]publicCommentResponse, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, newPublicCommentResponse(item))
		}

		response.OK(c, gin.H{
			"items":    items,
			"total":    page.Total,
			"page":     page.Page,
			"pageSize": page.PageSize,
		})
	}
}

func PublicCommentCount(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		site := middleware.GetSite(c)
		if site == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "site context missing"))
			return
		}

		var query publicCommentCountQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			writeBadRequest(c, "path is required")
			return
		}

		count, err := services.CountApprovedComments(site.ID, query.Path)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, gin.H{"count": count})
	}
}

func newPublicCommentResponse(node service.CommentNode) publicCommentResponse {
	result := publicCommentResponse{
		ID:            node.Comment.ID,
		ParentID:      node.Comment.ParentID,
		AuthorName:    node.Comment.AuthorName,
		AuthorWebsite: node.Comment.AuthorWebsite,
		Content:       node.Comment.Content,
		IsAdmin:       node.Comment.IsAdmin,
		CreatedAt:     node.Comment.CreatedAt,
	}
	if len(node.Replies) > 0 {
		result.Replies = make([]publicCommentResponse, 0, len(node.Replies))
		for _, reply := range node.Replies {
			result.Replies = append(result.Replies, newPublicCommentResponse(reply))
		}
	}
	return result
}
