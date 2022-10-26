package app

import (
	"internal/handler"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
}

func New() (*App, error) {
	srv := &http.Server{
		Addr: "localhost:8080",
	}
	return &App{
		HTTPServer: srv,
	}, nil
}

func (app *App) Run() error {
	http.Handle("/", http.HandlerFunc(handler.HandleURL)) // too many handle =|
	rand.Seed(time.Now().UnixNano())
	return app.HTTPServer.ListenAndServe()
}
