package app

import (
	"context"
	"github.com/Spear5030/yapshrtnr/db/migrate"
	"github.com/Spear5030/yapshrtnr/internal/config"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/Spear5030/yapshrtnr/internal/router"
	"github.com/Spear5030/yapshrtnr/internal/storage"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
}

func New(cfg config.Config) (*App, error) {

	var s interface {
		SetURL(user, short, long string)
		GetURL(short string) string
		GetURLsByUser(user string) (urls map[string]string)
		SetBatchURLs(ctx context.Context, urls []domain.URL) error
		Ping() error
	}
	if len(cfg.Database) > 0 {
		err := migrate.Migrate(cfg.Database, migrate.Migrations)
		if err != nil {
			return nil, err
		}
		pgStorage, err := storage.NewPGXStorage(cfg.Database)
		if err != nil {
			log.Fatal(err)
		}
		s = pgStorage
	} else if len(cfg.FileStorage) > 0 {
		fileStorage, err := storage.NewFileStorage(cfg.FileStorage)
		if err != nil {
			log.Fatal(err)
		}
		s = fileStorage
	} else {
		memoryStorage := storage.NewMemoryStorage()
		s = memoryStorage
	}
	h := handler.New(s, cfg.BaseURL)
	r := router.New(h)
	srv := &http.Server{
		Addr:    cfg.Addr,
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
