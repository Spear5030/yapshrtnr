package router

import (
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	//r.Use(handler.DecompressGZRequest)
	r.Get("/{id}", h.GetURL)
	r.Post("/", h.PostURL)

	r.Group(func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Post("/api/shorten", h.PostJSON)
	})

	return r
}
