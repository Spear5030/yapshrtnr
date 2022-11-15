package config

import "github.com/caarlos0/env"

type Config struct {
	Addr        string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
}

func New() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
