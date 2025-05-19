package adapters

import (
	"encoding/json"
	"net/http"
)

// writeErrorJSON writes a JSON error response.
func writeErrorJSON(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    status,
		"message": msg,
	})
}
