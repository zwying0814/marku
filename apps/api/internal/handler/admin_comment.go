package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/model"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type adminCommentsQuery struct {
	Status   string `form:"status"`
	Path     string `form:"path"`
	Q        string `form:"q"`
	From     string `form:"from"`
	To       string `form:"to"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Sort     string `form:"sort"`
}

type updateCommentRequest struct {
	Status string `json:"status" binding:"required"`
}

type replyCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type bulkCommentRequest struct {
	IDs    []uint `json:"ids" binding:"required"`
	Action string `json:"action" binding:"required"`
}

type adminCommentResponse struct {
	ID            uint      `json:"id"`
	SiteID        uint      `json:"siteId"`
	Path          string    `json:"path"`
	ParentID      *uint     `json:"parentId"`
	AuthorName    string    `json:"authorName"`
	AuthorEmail   string    `json:"authorEmail"`
	AuthorWebsite string    `json:"authorWebsite"`
	Content       string    `json:"content"`
	Status        string    `json:"status"`
	IsAdmin       bool      `json:"isAdmin"`
	IPHash        string    `json:"ipHash"`
	UserAgent     string    `json:"userAgent"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func ListAdminComments(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID, ok := parseSiteID(c)
		if !ok {
			return
		}

		var query adminCommentsQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			writeBadRequest(c, "invalid query parameters")
			return
		}
		page, pageSize := normalizeQueryPage(query.Page, query.PageSize)
		from, to, err := parseCommentDateRange(query.From, query.To)
		if err != nil {
			writeBadRequest(c, err.Error())
			return
		}

		comments, total, err := services.ListAdminComments(service.AdminCommentFilters{
			SiteID:   siteID,
			Status:   query.Status,
			Path:     query.Path,
			Search:   query.Q,
			From:     from,
			To:       to,
			Page:     page,
			PageSize: pageSize,
			Sort:     query.Sort,
		})
		if err != nil {
			writeServiceError(c, err)
			return
		}

		items := make([]adminCommentResponse, 0, len(comments))
		for index := range comments {
			items = append(items, newAdminCommentResponse(&comments[index]))
		}

		response.OK(c, gin.H{
			"items":    items,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func GetSiteCommentStats(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID, ok := parseSiteID(c)
		if !ok {
			return
		}
		days, err := parseRangeDays(c.Query("range"))
		if err != nil {
			writeBadRequest(c, err.Error())
			return
		}
		stats, err := services.GetSiteCommentStats(siteID, days)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, stats)
	}
}

func parseCommentDateRange(rawFrom, rawTo string) (*time.Time, *time.Time, error) {
	var from *time.Time
	var to *time.Time

	if rawFrom != "" {
		value, err := time.Parse("2006-01-02", rawFrom)
		if err != nil {
			return nil, nil, errors.New("from must be a date in YYYY-MM-DD format")
		}
		start := value.UTC()
		from = &start
	}
	if rawTo != "" {
		value, err := time.Parse("2006-01-02", rawTo)
		if err != nil {
			return nil, nil, errors.New("to must be a date in YYYY-MM-DD format")
		}
		end := value.UTC().AddDate(0, 0, 1)
		to = &end
	}
	if from != nil && to != nil && !from.Before(*to) {
		return nil, nil, errors.New("from must be before to")
	}
	return from, to, nil
}

func UpdateCommentStatus(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, ok := parseCommentID(c)
		if !ok {
			return
		}

		var request updateCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		comment, err := services.UpdateCommentStatus(commentID, request.Status)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, newAdminCommentResponse(comment))
	}
}

func DeleteComment(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, ok := parseCommentID(c)
		if !ok {
			return
		}
		if err := services.DeleteComment(commentID); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func ReplyComment(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, ok := parseCommentID(c)
		if !ok {
			return
		}
		admin := middleware.GetAdmin(c)
		if admin == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "admin context missing"))
			return
		}

		var request replyCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		comment, err := services.ReplyComment(commentID, admin, request.Content)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.Created(c, newAdminCommentResponse(comment))
	}
}

func BulkCommentAction(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID, ok := parseSiteID(c)
		if !ok {
			return
		}

		var request bulkCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		updated, err := services.BulkCommentAction(siteID, request.IDs, request.Action)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		response.OK(c, gin.H{"updated": updated})
	}
}

func newAdminCommentResponse(comment *model.Comment) adminCommentResponse {
	return adminCommentResponse{
		ID:            comment.ID,
		SiteID:        comment.SiteID,
		Path:          comment.Path,
		ParentID:      comment.ParentID,
		AuthorName:    comment.AuthorName,
		AuthorEmail:   comment.AuthorEmail,
		AuthorWebsite: comment.AuthorWebsite,
		Content:       comment.Content,
		Status:        comment.Status,
		IsAdmin:       comment.IsAdmin,
		IPHash:        comment.IPHash,
		UserAgent:     comment.UserAgent,
		CreatedAt:     comment.CreatedAt,
		UpdatedAt:     comment.UpdatedAt,
	}
}

func parseCommentID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		writeBadRequest(c, "invalid comment id")
		return 0, false
	}
	return uint(value), true
}

func normalizeQueryPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
