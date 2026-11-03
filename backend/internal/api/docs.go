package api

import (
	_ "embed"
	"io/fs"
	"net/http"

	swaggerFiles "github.com/swaggo/files/v2"

	"github.com/go-chi/chi/v5"
)

//go:embed openapi.yaml
var openAPISpec []byte

// Swagger UI assets are served from the embedded github.com/swaggo/files/v2
// bundle (pinned in go.mod/go.sum) instead of unpkg.com — the admin docs page
// no longer depends on a third-party CDN at runtime.
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>RioNexGate API</title>
  <link rel="stylesheet" href="/api/docs/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="/api/docs/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: '/api/openapi.yaml',
        dom_id: '#swagger-ui',
        persistAuthorization: true,
        presets: [SwaggerUIBundle.presets.apis],
        requestInterceptor: function(req) {
          var key = localStorage.getItem('rionexgate_api_key');
          if (key) req.headers['X-API-Key'] = key;
          return req;
        }
      });
    };
  </script>
</body>
</html>`

func (h *Handler) OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(openAPISpec)
}

func (h *Handler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerUIHTML))
}

// SwaggerUIAsset serves the embedded Swagger UI assets (CSS/JS); nothing is
// fetched from third-party CDNs.
func (h *Handler) SwaggerUIAsset(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "asset")
	var contentType string
	switch name {
	case "swagger-ui.css":
		contentType = "text/css; charset=utf-8"
	case "swagger-ui-bundle.js":
		contentType = "application/javascript; charset=utf-8"
	default:
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	data, err := fs.ReadFile(swaggerFiles.FS, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

func (h *Handler) ListProtocols(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]string{
		"protocols": {"vless", "vmess", "trojan"},
	})
}
