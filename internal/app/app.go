package app

import (
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/config"
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

func New(cfg config.Config) (*App, error) {

	s := storage.New()
	h := handler.New(s, fmt.Sprintf("%s:%d", cfg.Host, cfg.AppPort))
	r := router.New(h)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.AppPort),
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
