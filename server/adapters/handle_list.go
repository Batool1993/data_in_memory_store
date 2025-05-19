// server/adapters/handlers.go
package adapters

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

// pushList handles POST /v1/list/{key}/push.
func (h *Handlers) pushList(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	key := mux.Vars(req)["key"]
	if key == "" {
		writeErrorJSON(w, http.StatusBadRequest, "invalid key")
		return
	}

	var body listRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.storeService.LPush(req.Context(), key, body.Items...); err != nil {
		status := http.StatusBadRequest
		if !isClientError(err) {
			status = http.StatusInternalServerError
		}
		writeErrorJSON(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// popList handles POST /v1/list/{key}/pop.
func (h *Handlers) popList(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	key := mux.Vars(req)["key"]
	if key == "" {
		writeErrorJSON(w, http.StatusBadRequest, "invalid key")
		return
	}

	value, err := h.storeService.RPop(req.Context(), key)
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
