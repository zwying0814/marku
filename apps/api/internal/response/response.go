package response

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Code int

const (
	CodeOK                   Code = 0
	CodeBadRequest           Code = 40000
	CodeUnauthorized         Code = 40100
	CodeForbidden            Code = 40300
	CodeNotFound             Code = 40400
	CodeMethodNotAllowed     Code = 40500
	CodeConflict             Code = 40900
	CodePayloadTooLarge      Code = 41300
	CodeUnsupportedMediaType Code = 41500
	CodeRateLimited          Code = 42900
	CodeInternal             Code = 50000
	CodeUnavailable          Code = 50300
)

type Envelope struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type AppError struct {
	HTTPStatus int
	Code       Code
	Message    string
	Cause      error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(status int, code Code, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
	}
}

func Wrap(cause error, status int, code Code, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
		Cause:      cause,
	}
}

func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(err, http.StatusInternalServerError, CodeInternal, "internal server error")
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	})
}

func Error(c *gin.Context, status int, code Code, message string) {
	c.JSON(status, Envelope{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func Abort(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err)
	}

	appErr := AsAppError(err)
	if appErr == nil {
		appErr = New(http.StatusInternalServerError, CodeInternal, "internal server error")
	}
	c.AbortWithStatusJSON(appErr.HTTPStatus, Envelope{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}
