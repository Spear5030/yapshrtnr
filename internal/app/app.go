// Package app запускает само приложение
package app

import (
	"context"
	"go.uber.org/zap"
	"log"
	"net/http"

	"github.com/Spear5030/yapshrtnr/db/migrate"
	"github.com/Spear5030/yapshrtnr/internal/config"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/Spear5030/yapshrtnr/internal/router"
	"github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/Spear5030/yapshrtnr/pkg/logger"
)

// App основная структура приложения. HTTP сервер и логгер
type App struct {
	HTTPServer *http.Server
	logger     *zap.Logger
	tls        bool
}

// New возвращает App
func New(cfg config.Config) (*App, error) {

	var storager interface {
		SetURL(ctx context.Context, user, short, long string) error
		GetURL(ctx context.Context, short string) (string, bool)
		GetURLsByUser(ctx context.Context, user string) (urls map[string]string)
		SetBatchURLs(ctx context.Context, urls []domain.URL) error
		DeleteURLs(ctx context.Context, user string, shorts []string)
		//		Ping() error
	}

	lg, err := logger.New(true)
	if err != nil {
		return nil, err
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
		storager = pgStorage
		lg.Info("PostgreSQL storage.", zap.String("config", cfg.Database))
	} else if len(cfg.FileStorage) > 0 {
		fileStorage, err := storage.NewFileStorage(cfg.FileStorage)
		if err != nil {
			log.Fatal(err)
		}
		storager = fileStorage
	} else {
		memoryStorage := storage.NewMemoryStorage()
		storager = memoryStorage
		lg.Info("Inmemory storage.")
	}
	h := handler.New(lg, storager, cfg.BaseURL, cfg.Key)
	r := router.New(h)
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
	return &App{
		HTTPServer: srv,
		logger:     lg,
		tls:        cfg.HTTPS,
	}, nil
}

// Run запуск приложения.
func (app *App) Run() error {
	if app.tls {
		app.logger.Info("Listen with TLS")
		return app.HTTPServer.ListenAndServeTLS("cert/cert.pem", "cert/private.key")
	}
	app.logger.Info("Listen without TLS")
	return app.HTTPServer.ListenAndServe()
}
