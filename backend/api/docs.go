package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Email Platform API</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
<style>body { margin: 0; }</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/openapi.yaml",
    dom_id: "#swagger-ui",
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  });
};
</script>
</body>
</html>`

// RegisterDevRoutes registers development-only documentation routes.
func RegisterDevRoutes(router *gin.Engine, specPath string) {
	router.GET("/docs", docsHandler())
	router.GET("/openapi.yaml", openAPISpecHandler(specPath))
}

func docsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Header("Content-Security-Policy",
			"default-src 'self' https://unpkg.com; "+
				"script-src 'self' 'unsafe-inline' https://unpkg.com; "+
				"style-src 'self' 'unsafe-inline' https://unpkg.com; "+
				"img-src 'self' data: https://unpkg.com")
		c.String(http.StatusOK, swaggerHTML)
	}
}

func openAPISpecHandler(specPath string) gin.HandlerFunc {
	if specPath == "" {
		specPath = "api/openapi.yaml"
	}
	return func(c *gin.Context) {
		// #nosec G304 -- The path comes from trusted configuration, not request input.
		data, err := os.ReadFile(specPath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", data)
	}
}
