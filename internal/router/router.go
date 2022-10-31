package router

import (
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/{id}", h.GetURL)
	r.Post("/", h.PostURL)
	return r
}
