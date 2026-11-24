package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"marku/api/internal/response"
	"marku/api/internal/service"
)

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Abort(c, response.New(http.StatusUnauthorized, response.CodeUnauthorized, service.ErrInvalidCredentials.Error()))
	case errors.Is(err, service.ErrInvalidToken):
		response.Abort(c, response.New(http.StatusUnauthorized, response.CodeUnauthorized, service.ErrInvalidToken.Error()))
	case errors.Is(err, service.ErrInvalidSiteKey):
		response.Abort(c, response.New(http.StatusUnauthorized, response.CodeUnauthorized, service.ErrInvalidSiteKey.Error()))
	case errors.Is(err, service.ErrAuthNotConfigured):
		response.Abort(c, response.New(http.StatusServiceUnavailable, response.CodeUnavailable, service.ErrAuthNotConfigured.Error()))
	case errors.Is(err, service.ErrAdminDisabled):
		response.Abort(c, response.New(http.StatusForbidden, response.CodeForbidden, service.ErrAdminDisabled.Error()))
	case errors.Is(err, service.ErrSiteDisabled):
		response.Abort(c, response.New(http.StatusForbidden, response.CodeForbidden, service.ErrSiteDisabled.Error()))
	case errors.Is(err, service.ErrInvalidCurrentPassword):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, service.ErrInvalidCurrentPassword.Error()))
	case errors.Is(err, service.ErrWeakPassword):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, service.ErrWeakPassword.Error()))
	case errors.Is(err, service.ErrCommentNotFound):
		response.Abort(c, response.New(http.StatusNotFound, response.CodeNotFound, service.ErrCommentNotFound.Error()))
	case errors.Is(err, service.ErrCommentTooFast):
		c.Header("Retry-After", "5")
		response.Abort(c, response.New(http.StatusTooManyRequests, response.CodeRateLimited, service.ErrCommentTooFast.Error()))
	case errors.Is(err, service.ErrCommentRateLimited):
		c.Header("Retry-After", "60")
		response.Abort(c, response.New(http.StatusTooManyRequests, response.CodeRateLimited, service.ErrCommentRateLimited.Error()))
	case errors.Is(err, service.ErrDuplicateComment):
		response.Abort(c, response.New(http.StatusConflict, response.CodeConflict, service.ErrDuplicateComment.Error()))
	case errors.Is(err, service.ErrInvalidParent), errors.Is(err, service.ErrCommentDepthExceeded), errors.Is(err, service.ErrInvalidCommentAction):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, err.Error()))
	case errors.Is(err, service.ErrInvalidCommentInput):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, err.Error()))
	case errors.Is(err, service.ErrInvalidPagePath), errors.Is(err, service.ErrInvalidViewInput):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, err.Error()))
	case errors.Is(err, service.ErrInvalidLikeInput):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, err.Error()))
	case errors.Is(err, service.ErrSlugExists):
		response.Abort(c, response.New(http.StatusConflict, response.CodeConflict, service.ErrSlugExists.Error()))
	case errors.Is(err, service.ErrInvalidSiteInput):
		response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, err.Error()))
	case errors.Is(err, service.ErrSiteNotFound):
		response.Abort(c, response.New(http.StatusNotFound, response.CodeNotFound, service.ErrSiteNotFound.Error()))
	default:
		response.Abort(c, response.Wrap(err, http.StatusInternalServerError, response.CodeInternal, "internal server error"))
	}
}

func writeBadRequest(c *gin.Context, message string) {
	response.Abort(c, response.New(http.StatusBadRequest, response.CodeBadRequest, message))
}
