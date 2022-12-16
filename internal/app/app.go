package app

import (
	"context"
	"github.com/Spear5030/yapshrtnr/db/migrate"
	"github.com/Spear5030/yapshrtnr/internal/config"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/Spear5030/yapshrtnr/internal/handler"
	"github.com/Spear5030/yapshrtnr/internal/router"
	"github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/Spear5030/yapshrtnr/pkg/logger"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type App struct {
	HTTPServer *http.Server
	logger     *zap.Logger
}

func New(cfg config.Config) (*App, error) {

	var storager interface {
		SetURL(ctx context.Context, user, short, long string) error
		GetURL(ctx context.Context, short string) string
		GetURLsByUser(ctx context.Context, user string) (urls map[string]string)
		SetBatchURLs(ctx context.Context, urls []domain.URL) error
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
	} else if len(cfg.FileStorage) > 0 {
		fileStorage, err := storage.NewFileStorage(cfg.FileStorage)
		if err != nil {
			log.Fatal(err)
		}
		storager = fileStorage
	} else {
		memoryStorage := storage.NewMemoryStorage()
		storager = memoryStorage
	}
	h := handler.New(lg, storager, cfg.BaseURL, cfg.Key)
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
