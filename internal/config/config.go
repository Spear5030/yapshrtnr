// Package config определяет конфигурирование приложения.
package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// Config содержит строки конфигурации приложения. Значения собираются из ENV.
type Config struct {
	Addr        string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	Database    string `env:"DATABASE_DSN"`
	Key         string `env:"COOKIES_KEY" envDefault:"V3ry$trongK3y"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server Address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "path to file storage")
	flag.StringVar(&cfg.Database, "d", cfg.Database, "DSN for PGSQL")
	flag.StringVar(&cfg.Database, "k", cfg.Key, "Key string for sign cookies")
}

// New возвращает конфиг с дефолтными значениями, если не указаны флагами.
func New() (Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	flag.Parse()
	return cfg, nil
}
