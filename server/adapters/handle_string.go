// server/adapters/handlers.go
package adapters

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// setString handles POST /v1/string/{key}.
func (h *Handlers) setString(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	key := mux.Vars(req)["key"]
	if key == "" {
		writeErrorJSON(w, http.StatusBadRequest, "invalid key")
		return
	}

	var body stringRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	log.Printf("[Handler] raw TTLSeconds=%d (from JSON)", body.TTLSeconds)
	ttl := time.Duration(body.TTLSeconds) * time.Second

	if err := h.storeService.SetString(req.Context(), key, body.Value, ttl); err != nil {
		status := http.StatusBadRequest
		if !isClientError(err) {
			status = http.StatusInternalServerError
		}
		writeErrorJSON(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// getString handles GET /v1/string/{key}.
func (h *Handlers) getString(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	key := mux.Vars(req)["key"]
	if key == "" {
		writeErrorJSON(w, http.StatusBadRequest, "invalid key")
		return
	}

	value, err := h.storeService.GetString(req.Context(), key)
	if err != nil {
		status := http.StatusBadRequest
		if !isClientError(err) {
			status = http.StatusInternalServerError
		}
		writeErrorJSON(w, status, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"value": value})
}

// deleteString handles DELETE /v1/string/{key}.
func (h *Handlers) deleteString(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	key := mux.Vars(req)["key"]
	if key == "" {
		writeErrorJSON(w, http.StatusBadRequest, "invalid key")
		return
	}

	if err := h.storeService.DeleteString(req.Context(), key); err != nil {
		status := http.StatusBadRequest
		if !isClientError(err) {
			status = http.StatusInternalServerError
		}
		writeErrorJSON(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
