// Package config определяет конфигурирование приложения.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
)

// Config содержит строки конфигурации приложения. Значения собираются из ENV.
type Config struct {
	Addr        string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL     string `env:"BASE_URL" json:"base_url"`
	FileStorage string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Database    string `env:"DATABASE_DSN" json:"database_dsn"`
	Key         string `env:"COOKIES_KEY" envDefault:"V3ry$trongK3y"`
	HTTPS       bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	Config      string `env:"CONFIG"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server Address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "path to file storage")
	flag.StringVar(&cfg.Database, "d", cfg.Database, "DSN for PGSQL")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Key string for sign cookies")
	flag.BoolVar(&cfg.HTTPS, "s", cfg.HTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "Config file destination")
	flag.StringVar(&cfg.Config, "config", cfg.Config, "Config file destination")
}

// New возвращает конфиг. Приоритет file->env->flag
func New() (Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	flag.Parse()
	if len(cfg.Config) > 0 {
		err := ReadConfig(cfg.Config)
		if err != nil {
			log.Println(err)
		}
		_ = env.Parse(&cfg) //если ошибка есть, ее отловили выше
		flag.Parse()        //повторные вызовы для приоритизации по ТЗ file->env->flag
	}
	return cfg, nil
}

// ReadConfig читает из файла конфиг в JSON-формате.
func ReadConfig(fname string) error {
	cfgFile, err := os.ReadFile(fname)
	if err != nil {
		log.Println("Read Config error:", err)
		return err
	}
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Println("Read Config error:", err)
		return err
	}
	return nil
}
