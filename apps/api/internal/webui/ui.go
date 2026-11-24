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

//go:embed all:docs-dist
var docsAssets embed.FS

type UI struct {
	admin          fs.FS
	docs           fs.FS
	adminAvailable bool
	docsAvailable  bool
}

func Default() *UI {
	admin, _ := fs.Sub(adminAssets, "admin-dist")
	docs, _ := fs.Sub(docsAssets, "docs-dist")
	return New(admin, docs)
}

func New(admin, docs fs.FS) *UI {
	return &UI{
		admin:          admin,
		docs:           docs,
		adminAvailable: fileExists(admin, "index.html"),
		docsAvailable:  fileExists(docs, "index.html"),
	}
}

func (u *UI) Register(engine *gin.Engine) {
	if !u.docsAvailable {
		return
	}
	engine.GET("/docs", redirectToDocs)
	engine.HEAD("/docs", redirectToDocs)
	engine.GET("/docs/*filepath", u.serveDocs)
	engine.HEAD("/docs/*filepath", u.serveDocs)
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

func redirectToDocs(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, "/docs/")
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

func (u *UI) serveDocs(c *gin.Context) {
	rel, ok := cleanPath(c.Param("filepath"))
	if !ok {
		http.NotFound(c.Writer, c.Request)
		return
	}
	if rel == "" {
		if serveFile(c, u.docs, "index.html", http.StatusOK) {
			return
		}
		u.serveDocsNotFound(c)
		return
	}

	if serveFile(c, u.docs, rel, http.StatusOK) {
		return
	}
	if path.Ext(rel) == "" {
		if serveFile(c, u.docs, rel+".html", http.StatusOK) {
			return
		}
		if serveFile(c, u.docs, path.Join(rel, "index.html"), http.StatusOK) {
			return
		}
	}
	u.serveDocsNotFound(c)
}

func (u *UI) serveDocsNotFound(c *gin.Context) {
	if serveFile(c, u.docs, "404.html", http.StatusNotFound) {
		return
	}
	http.NotFound(c.Writer, c.Request)
}

func serveFile(c *gin.Context, fsys fs.FS, name string, status int) bool {
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
	switch {
	case strings.HasSuffix(name, ".html"):
		c.Header("Cache-Control", "no-cache")
	case strings.HasPrefix(name, "assets/"), strings.HasPrefix(name, "static/"):
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	default:
		c.Header("Cache-Control", "public, max-age=3600")
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
		strings.HasPrefix(raw, "/swagger/")
}
