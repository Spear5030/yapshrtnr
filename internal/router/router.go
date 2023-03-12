// Package router реализует роутер. Связь эндпоинтов с обработчиками
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Spear5030/yapshrtnr/internal/handler"
)

// New возвращает роутер с группами нужных эндпоинтов.
func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(handler.CheckCookies(h.SecretKey))
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(handler.DecompressGZRequest)
	r.Mount("/debug", middleware.Profiler())
	r.Get("/{id}", h.GetURL)
	r.Get("/ping", h.PingDB)
	r.Post("/", h.PostURL)
	r.Get("/api/internal/stats", h.GetInternalStats)

	r.Group(func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Post("/api/shorten", h.PostJSON)
		r.Get("/api/user/urls", h.GetURLsByUser)
		r.Delete("/api/user/urls", h.DeleteBatchByUser)
		r.Post("/api/shorten/batch", h.PostBatch)
	})

	return r
}
