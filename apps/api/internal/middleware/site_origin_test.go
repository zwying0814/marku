package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"marku/api/internal/model"
)

func TestSiteOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	site := &model.Site{Name: "Test"}
	if err := site.SetAllowedOrigins([]string{"https://blog.example.com"}); err != nil {
		t.Fatalf("set origins: %v", err)
	}

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(siteContextKey, site)
		c.Next()
	}, SiteOrigin())
	engine.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{name: "allowed write", method: http.MethodPost, origin: "https://blog.example.com", want: http.StatusOK},
		{name: "denied write", method: http.MethodPost, origin: "https://evil.example.com", want: http.StatusForbidden},
		{name: "server to server write", method: http.MethodPost, want: http.StatusOK},
		{name: "read is not origin restricted", method: http.MethodGet, origin: "https://evil.example.com", want: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/test", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
		})
	}
}
