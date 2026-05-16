package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed sdk_api.json
var openAPISpec string

// DocsHandler Swagger UI 文档处理器
type DocsHandler struct{}

// NewDocsHandler 创建 DocsHandler
func NewDocsHandler() *DocsHandler { return &DocsHandler{} }

// ServeUI 处理 GET /docs — 返回 Swagger UI HTML
func (h *DocsHandler) ServeUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerUIHTML)
}

// ServeSpec 处理 GET /docs/openapi.json — 返回 OpenAPI 3.0 规范
func (h *DocsHandler) ServeSpec(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	c.String(http.StatusOK, openAPISpec)
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AdLab Server — SDK API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { background: #111827; padding: 8px 20px; }
    .swagger-ui .topbar-wrapper .link { display: flex; align-items: center; gap: 10px; }
    .swagger-ui .topbar-wrapper .link::before {
      content: 'AdLab SDK API';
      color: #e8612c;
      font-size: 18px;
      font-weight: 700;
      letter-spacing: 0.3px;
    }
    .swagger-ui .topbar .download-url-wrapper { display: none; }
    .swagger-ui .info .title { color: #111827; font-size: 28px; }
    .swagger-ui .info .description p { color: #374151; }
    .swagger-ui .opblock.opblock-get .opblock-summary-method { background: #2563eb; }
    .swagger-ui .opblock.opblock-post .opblock-summary-method { background: #059669; }
    .swagger-ui .btn.execute { background: #e8612c; border-color: #e8612c; color: #fff; }
    .swagger-ui .btn.execute:hover { background: #d4521f; }
    .swagger-ui section.models { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '/docs/openapi.json',
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: 'BaseLayout',
      deepLinking: true,
      defaultModelsExpandDepth: -1,
      defaultModelExpandDepth: 2,
      tryItOutEnabled: true,
      displayRequestDuration: true,
      filter: true,
      tagsSorter: 'alpha',
    });
  </script>
</body>
</html>`
