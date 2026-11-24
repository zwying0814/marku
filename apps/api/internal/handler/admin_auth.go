package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"marku/api/internal/middleware"
	"marku/api/internal/model"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type adminResponse struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

func AdminLogin(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request loginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		admin, token, expiresAt, err := services.Login(request.Email, request.Password)
		if err != nil {
			writeServiceError(c, err)
			return
		}

		response.OK(c, gin.H{
			"token":     token,
			"expiresAt": expiresAt,
			"admin":     newAdminResponse(admin),
		})
	}
}

func AdminMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := middleware.GetAdmin(c)
		if admin == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "admin context missing"))
			return
		}
		response.OK(c, newAdminResponse(admin))
	}
}

func AdminChangePassword(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request changePasswordRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeBadRequest(c, "invalid request body")
			return
		}

		admin := middleware.GetAdmin(c)
		if admin == nil {
			response.Abort(c, response.New(http.StatusInternalServerError, response.CodeInternal, "admin context missing"))
			return
		}
		if err := services.ChangePassword(admin, request.CurrentPassword, request.NewPassword); err != nil {
			writeServiceError(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func newAdminResponse(admin *model.Admin) adminResponse {
	return adminResponse{
		ID:          admin.ID,
		Email:       admin.Email,
		Nickname:    admin.Nickname,
		Status:      admin.Status,
		LastLoginAt: admin.LastLoginAt,
		CreatedAt:   admin.CreatedAt,
		UpdatedAt:   admin.UpdatedAt,
	}
}
