package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"marku/api/internal/config"
	"marku/api/internal/handler"
	"marku/api/internal/middleware"
	"marku/api/internal/response"
	"marku/api/internal/service"
	"marku/api/internal/version"
	"marku/api/internal/webui"
)

type Option func(*routerOptions)

type routerOptions struct {
	webUI        *webui.UI
	disableWebUI bool
}

func WithoutWebUI() Option {
	return func(options *routerOptions) {
		options.disableWebUI = true
	}
}

func New(cfg config.Config, db *gorm.DB, services *service.Service, options ...Option) (*gin.Engine, error) {
	settings := routerOptions{}
	for _, option := range options {
		option(&settings)
	}

	switch cfg.Env {
	case config.EnvProduction:
		gin.SetMode(gin.ReleaseMode)
	case config.EnvTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.HandleMethodNotAllowed = true

	trustedProxies := cfg.HTTP.TrustedProxies
	if len(trustedProxies) == 0 {
		trustedProxies = nil
	}
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	engine.Use(
		middleware.RequestID(),
		middleware.AccessLogger(),
		middleware.Recovery(),
		middleware.CORS(cfg.CORS),
		middleware.BodyLimit(cfg.HTTP.MaxBodyBytes),
	)

	engine.GET("/healthz", handler.Health)
	engine.GET("/readyz", handler.Ready(db))
	engine.GET("/openapi.yaml", handler.OpenAPISpec)
	engine.GET("/swagger", handler.SwaggerUI)
	engine.GET("/swagger/", handler.SwaggerUI)

	var embeddedUI *webui.UI
	if !settings.disableWebUI {
		embeddedUI = settings.webUI
		if embeddedUI == nil {
			embeddedUI = webui.Default()
		}
		embeddedUI.Register(engine)
	}

	api := engine.Group("/api/v1")
	api.Use(middleware.RateLimit(cfg.RateLimit))
	api.GET("", func(c *gin.Context) {
		response.OK(c, gin.H{
			"service": "marku-api",
			"version": version.Current,
		})
	})

	adminPublic := api.Group("/admin")
	adminPublic.POST("/auth/login", handler.AdminLogin(services))

	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuth(services))
	admin.GET("/auth/me", handler.AdminMe())
	admin.POST("/auth/change-password", handler.AdminChangePassword(services))
	admin.GET("/overview", handler.GetAdminOverview(services))
	admin.GET("/sites", handler.ListSites(services))
	admin.POST("/sites", handler.CreateSite(services))
	admin.GET("/sites/:id", handler.GetSite(services))
	admin.PATCH("/sites/:id", handler.UpdateSite(services))
	admin.DELETE("/sites/:id", handler.DeleteSite(services))
	admin.POST("/sites/:id/rotate-key", handler.RotateSiteKey(services))
	admin.DELETE("/sites/:id/key", handler.RevokeSiteKey(services))
	admin.GET("/sites/:id/comments", handler.ListAdminComments(services))
	admin.GET("/sites/:id/comments/stats", handler.GetSiteCommentStats(services))
	admin.POST("/sites/:id/comments/bulk", handler.BulkCommentAction(services))
	admin.GET("/sites/:id/stats", handler.GetSiteViewStats(services))
	admin.GET("/sites/:id/likes", handler.GetSiteLikeStats(services))
	admin.PATCH("/comments/:id", handler.UpdateCommentStatus(services))
	admin.DELETE("/comments/:id", handler.DeleteComment(services))
	admin.POST("/comments/:id/reply", handler.ReplyComment(services))

	public := api.Group("/public")
	public.Use(middleware.SiteKey(services), middleware.SiteOrigin())
	public.GET("/site", handler.PublicSite())
	public.GET("/comments", handler.ListPublicComments(services))
	public.POST("/comments", handler.CreatePublicComment(services))
	public.GET("/comments/count", handler.PublicCommentCount(services))
	public.POST("/views", handler.RecordPublicView(services))
	public.GET("/views", handler.GetPublicViews(services))
	public.POST("/likes", handler.TogglePublicLike(services))
	public.PUT("/likes", handler.SetPublicLike(services))
	public.DELETE("/likes", handler.UnsetPublicLike(services))
	public.GET("/likes", handler.GetPublicLike(services))
	public.GET("/stats", handler.GetPublicPageStats(services))

	engine.NoRoute(func(c *gin.Context) {
		if embeddedUI != nil && embeddedUI.HandleNoRoute(c) {
			return
		}
		response.Abort(c, response.New(http.StatusNotFound, response.CodeNotFound, "not found"))
	})
	engine.NoMethod(func(c *gin.Context) {
		response.Abort(c, response.New(http.StatusMethodNotAllowed, response.CodeMethodNotAllowed, "method not allowed"))
	})

	return engine, nil
}
