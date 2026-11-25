package webui

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:admin-dist
var adminAssets embed.FS

//go:embed all:embed-dist
var embedAssets embed.FS

type UI struct {
	admin          fs.FS
	embed          fs.FS
	adminAvailable bool
	embedAvailable bool
}

func Default() *UI {
	admin, _ := fs.Sub(adminAssets, "admin-dist")
	embedFS, _ := fs.Sub(embedAssets, "embed-dist")
	return New(admin, embedFS)
}

func New(admin fs.FS, embedFS ...fs.FS) *UI {
	var embedAssets fs.FS
	if len(embedFS) > 0 {
		embedAssets = embedFS[0]
	}
	return &UI{
		admin:          admin,
		embed:          embedAssets,
		adminAvailable: fileExists(admin, "index.html"),
		embedAvailable: fileExists(embedAssets, "marku-embed.js"),
	}
}

func (u *UI) Register(engine *gin.Engine) {
	if !u.embedAvailable {
		return
	}
	engine.GET("/embed/v1/*filepath", u.serveEmbed)
	engine.HEAD("/embed/v1/*filepath", u.serveEmbed)
}

func (u *UI) serveEmbed(c *gin.Context) {
	rel, ok := cleanPath(c.Param("filepath"))
	if !ok || rel == "" {
		http.NotFound(c.Writer, c.Request)
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	if serveFileWithCache(c, u.embed, rel, http.StatusOK, "public, max-age=31536000, immutable") {
		return
	}
	http.NotFound(c.Writer, c.Request)
}

func (u *UI) HandleNoRoute(c *gin.Context) bool {
	if !u.adminAvailable || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
		return false
	}
	if reservedPath(c.Request.URL.Path) {
		return false
	}
	return u.serveAdmin(c)
}

func (u *UI) serveAdmin(c *gin.Context) bool {
	rel, ok := cleanPath(c.Request.URL.Path)
	if !ok {
		return false
	}
	if rel != "" {
		if serveFile(c, u.admin, rel, http.StatusOK) {
			return true
		}
		if path.Ext(rel) != "" {
			c.Status(http.StatusNotFound)
			return true
		}
	}
	return serveFile(c, u.admin, "index.html", http.StatusOK)
}

func serveFile(c *gin.Context, fsys fs.FS, name string, status int) bool {
	return serveFileWithCache(c, fsys, name, status, "")
}

func serveFileWithCache(c *gin.Context, fsys fs.FS, name string, status int, cacheControl string) bool {
	if fsys == nil || !fs.ValidPath(name) {
		return false
	}
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return false
	}

	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	if cacheControl != "" {
		c.Header("Cache-Control", cacheControl)
	} else {
		switch {
		case strings.HasSuffix(name, ".html"):
			c.Header("Cache-Control", "no-cache")
		case strings.HasPrefix(name, "assets/"), strings.HasPrefix(name, "static/"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		default:
			c.Header("Cache-Control", "public, max-age=3600")
		}
	}

	c.Status(status)
	if c.Request.Method != http.MethodHead {
		_, _ = c.Writer.Write(data)
	}
	return true
}

func cleanPath(raw string) (string, bool) {
	if strings.Contains(raw, "\\") {
		return "", false
	}
	clean := path.Clean("/" + strings.TrimPrefix(raw, "/"))
	if clean == "/" {
		return "", true
	}
	rel := strings.TrimPrefix(clean, "/")
	if rel == "" || strings.HasPrefix(rel, "../") || !fs.ValidPath(rel) {
		return "", false
	}
	return rel, true
}

func fileExists(fsys fs.FS, name string) bool {
	if fsys == nil {
		return false
	}
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func reservedPath(raw string) bool {
	switch raw {
	case "/api", "/docs", "/swagger", "/healthz", "/readyz", "/openapi.yaml":
		return true
	}
	return strings.HasPrefix(raw, "/api/") || strings.HasPrefix(raw, "/docs/") ||
		strings.HasPrefix(raw, "/swagger/") || strings.HasPrefix(raw, "/embed/")
}
