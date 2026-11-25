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
	embed := fstest.MapFS{
		"marku-embed.js":      &fstest.MapFile{Data: []byte("console.log('embed')")},
		"marku-embed.iife.js": &fstest.MapFile{Data: []byte("var MarkuEmbed;")},
	}
	ui := New(admin, embed)

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
		wantHeaders map[string]string
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
			name:       "docs path is not served by the binary",
			method:     http.MethodGet,
			target:     "/docs/guide/getting-started",
			wantStatus: http.StatusNotFound,
			wantBody:   "api not found",
		},
		{
			name:       "api route is not admin",
			method:     http.MethodGet,
			target:     "/api/v1/missing",
			wantStatus: http.StatusNotFound,
			wantBody:   "api not found",
		},
		{
			name:        "embed script",
			method:      http.MethodGet,
			target:      "/embed/v1/marku-embed.js",
			wantStatus:  http.StatusOK,
			wantBody:    "console.log('embed')",
			wantContent: "javascript",
			wantHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Cache-Control":               "public, max-age=31536000, immutable",
			},
		},
		{
			name:       "embed iife script",
			method:     http.MethodGet,
			target:     "/embed/v1/marku-embed.iife.js",
			wantStatus: http.StatusOK,
			wantBody:   "var MarkuEmbed;",
		},
		{
			name:       "missing embed asset",
			method:     http.MethodGet,
			target:     "/embed/v1/missing.js",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "embed path is not admin fallback",
			method:     http.MethodGet,
			target:     "/embed/v1/unknown",
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found",
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
			for header, want := range test.wantHeaders {
				if got := recorder.Header().Get(header); got != want {
					t.Fatalf("%s = %q, want %q", header, got, want)
				}
			}
		})
	}

}
