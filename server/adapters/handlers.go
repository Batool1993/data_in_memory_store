package adapters

import (
	"data_storage/server/adapters/middleware"
	"data_storage/server/store_service"
	"github.com/gorilla/mux"
	"net/http"
)

// Handlers holds references to the use-case service.
type Handlers struct {
	storeService store_service.StoreServiceRepo
}

// NewHandler constructs HTTP handlers from the service.
func NewHandler(s store_service.StoreServiceRepo, expectedToken string) http.Handler {
	h := &Handlers{storeService: s}
	router := mux.NewRouter()
	h.RegisterHandlers(router)
	router.Use(middleware.LoggingMiddleWare, middleware.RecoveryMiddleware, middleware.TokenAuth(expectedToken))

	return router
}

// RegisterHandlers wires handlers onto the mux.Router.
func (h *Handlers) RegisterHandlers(router *mux.Router) {
	router.HandleFunc("/v1/string/{key}", h.setString).Methods("POST")
	router.HandleFunc("/v1/string/{key}", h.getString).Methods("GET")
	router.HandleFunc("/v1/string/{key}", h.deleteString).Methods("DELETE")

	list := router.PathPrefix("/v1/list/{key}").Subrouter()
	list.HandleFunc("/push", h.pushList).Methods("POST")
	list.HandleFunc("/pop", h.popList).Methods("POST")
}
