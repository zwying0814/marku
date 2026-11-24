package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestEmbeddedUI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	admin := fstest.MapFS{
		"index.html":           &fstest.MapFile{Data: []byte("<html>admin</html>")},
		"assets/admin.1234.js": &fstest.MapFile{Data: []byte("console.log('admin')")},
		"favicon.svg":          &fstest.MapFile{Data: []byte("<svg></svg>")},
	}
	docs := fstest.MapFS{
		"index.html":                 &fstest.MapFile{Data: []byte("<html>docs</html>")},
		"guide/getting-started.html": &fstest.MapFile{Data: []byte("<html>started</html>")},
		"static/css/styles.1234.css": &fstest.MapFile{Data: []byte("body{}")},
		"404.html":                   &fstest.MapFile{Data: []byte("<html>docs 404</html>")},
	}
	ui := New(admin, docs)

	engine := gin.New()
	ui.Register(engine)
	engine.NoRoute(func(c *gin.Context) {
		if ui.HandleNoRoute(c) {
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"message": "api not found"})
	})

	tests := []struct {
		name        string
		method      string
		target      string
		wantStatus  int
		wantBody    string
		wantContent string
	}{
		{
			name:       "admin root",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			wantBody:   "<html>admin</html>",
		},
		{
			name:       "admin spa fallback",
			method:     http.MethodGet,
			target:     "/sites/1",
			wantStatus: http.StatusOK,
			wantBody:   "<html>admin</html>",
		},
		{
			name:        "admin hashed asset",
			method:      http.MethodGet,
			target:      "/assets/admin.1234.js",
			wantStatus:  http.StatusOK,
			wantBody:    "console.log('admin')",
			wantContent: "javascript",
		},
		{
			name:       "missing admin asset",
			method:     http.MethodGet,
			target:     "/assets/missing.js",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "docs root",
			method:     http.MethodGet,
			target:     "/docs/",
			wantStatus: http.StatusOK,
			wantBody:   "<html>docs</html>",
		},
		{
			name:       "docs clean route",
			method:     http.MethodGet,
			target:     "/docs/guide/getting-started",
			wantStatus: http.StatusOK,
			wantBody:   "<html>started</html>",
		},
		{
			name:       "docs static asset",
			method:     http.MethodGet,
			target:     "/docs/static/css/styles.1234.css",
			wantStatus: http.StatusOK,
			wantBody:   "body{}",
		},
		{
			name:       "docs not found",
			method:     http.MethodGet,
			target:     "/docs/missing",
			wantStatus: http.StatusNotFound,
			wantBody:   "<html>docs 404</html>",
		},
		{
			name:       "api route is not admin",
			method:     http.MethodGet,
			target:     "/api/v1/missing",
			wantStatus: http.StatusNotFound,
			wantBody:   "api not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.target, nil)
			engine.ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d (body=%s)", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if test.wantBody != "" && !strings.Contains(recorder.Body.String(), test.wantBody) {
				t.Fatalf("body = %q, want contains %q", recorder.Body.String(), test.wantBody)
			}
			if test.wantContent != "" && !strings.Contains(recorder.Header().Get("Content-Type"), test.wantContent) {
				t.Fatalf("content-type = %q, want contains %q", recorder.Header().Get("Content-Type"), test.wantContent)
			}
		})
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if recorder.Code != http.StatusPermanentRedirect || recorder.Header().Get("Location") != "/docs/" {
		t.Fatalf("docs redirect = status %d location %q", recorder.Code, recorder.Header().Get("Location"))
	}
}
