package api

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec []byte

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>RioNexGate API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
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

func (h *Handler) ListProtocols(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]string{
		"protocols": {"vless", "vmess", "trojan"},
	})
}
