package app

import (
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/Spear5030/yapshrtnr/internal/storage"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
	Handler    *handler.Handler
}

func New() (*App, error) {
	srv := &http.Server{
		Addr: "localhost:8080",
	}
	s := storage.NewStorage()
	h := handler.New(s)
	return &App{
		HTTPServer: srv,
		Handler:    h,
	}, nil
}

func (app *App) Run() error {
	//http.Handle("/", http.HandlerFunc(handler.HandleURL)) // too many handle =|
	http.Handle("/", app.Handler)
	rand.Seed(time.Now().UnixNano())
	return app.HTTPServer.ListenAndServe()
}
