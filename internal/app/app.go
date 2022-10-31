package app

import (
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/Spear5030/yapshrtnr/internal/router"
	"github.com/Spear5030/yapshrtnr/internal/storage"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
}

func New() (*App, error) {

	s := storage.NewStorage()
	h := handler.New(s)
	r := router.New(h)
	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: r,
	}
	return &App{
		HTTPServer: srv,
	}, nil
}

func (app *App) Run() error {
	rand.Seed(time.Now().UnixNano())
	return app.HTTPServer.ListenAndServe()
}
