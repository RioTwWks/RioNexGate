package api

import (
	"encoding/json"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"
)

// maxJSONBody caps JSON request bodies (DoS / memory abuse).
const maxJSONBody = 1 << 20 // 1 MiB

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func generateQR(content string) ([]byte, error) {
	return qrcode.Encode(content, qrcode.Medium, 256)
}
