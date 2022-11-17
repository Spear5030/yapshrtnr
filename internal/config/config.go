package config

import (
	"flag"
	"github.com/caarlos0/env"
)

type Config struct {
	Addr        string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server Address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "path to file storage")
}

func New() (Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	flag.Parse()
	return cfg, nil
}
