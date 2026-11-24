package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"marku/api/internal/docs"
)

func OpenAPISpec(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", docs.OpenAPI)
}

func SwaggerUI(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(docs.SwaggerHTML))
}
