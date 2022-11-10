package config

import "github.com/caarlos0/env"

type Config struct {
	AppPort int    `env:"APP_PORT" envDefault:"8080"`
	Host    string `env:"APP_HOST" envDefault:"localhost"`
}

func New() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
