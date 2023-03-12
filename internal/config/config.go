// Package config определяет конфигурирование приложения.
package config

import (
	"encoding/json"
	"flag"
	"github.com/caarlos0/env"
	"log"
	"net"
	"os"
)

const (
	defaultAddr    = "localhost:8080"
	defaultBaseURL = "http://localhost:8080"
)

// CustomIPNet кастомный net.IPNet для интрейфесов из flag, env,json
type CustomIPNet net.IPNet

// UnmarshalText  для получения net.IPNet из строки из ENV
func (t *CustomIPNet) UnmarshalText(data []byte) error {
	_, ipNet, err := net.ParseCIDR(string(data))
	if err != nil {
		log.Println("error parsing trusted subnet from ENV", err)
		return err
	}
	*t = CustomIPNet(*ipNet)
	return err
}

// UnmarshalJSON для получения net.IPNet из строки из JSON конфига
func (t *CustomIPNet) UnmarshalJSON(data []byte) error {
	_, ipNet, err := net.ParseCIDR(string(data))
	if err != nil {
		log.Println("error parsing trusted subnet from JSON", err)
		return err
	}
	*t = CustomIPNet(*ipNet)
	return err
}

// String для имплементации flag.Value interface - net.IPNet из строки из флага
func (t *CustomIPNet) String() string {
	return t.Mask.String()
}

// Set для имплементации flag.Value interface - net.IPNet из строки из флага
func (t *CustomIPNet) Set(data string) error {
	_, ipNet, err := net.ParseCIDR(data)
	if err != nil {
		log.Println("error parsing trusted subnet from flag", err)
		return err
	}
	*t = CustomIPNet(*ipNet)
	return err
}

// Config содержит строки конфигурации приложения. Значения собираются из ENV.
type Config struct {
	Addr          string      `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL       string      `env:"BASE_URL" json:"base_url"`
	FileStorage   string      `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Database      string      `env:"DATABASE_DSN" json:"database_dsn"`
	Key           string      `env:"COOKIES_KEY" envDefault:"V3ry$trongK3y"`
	HTTPS         bool        `env:"ENABLE_HTTPS" json:"enable_https"`
	Config        string      `env:"CONFIG"`
	TrustedSubnet CustomIPNet `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server Address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "path to file storage")
	flag.StringVar(&cfg.Database, "d", cfg.Database, "DSN for PGSQL")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Key string for sign cookies")
	flag.BoolVar(&cfg.HTTPS, "s", cfg.HTTPS, "Enable HTTPS")
	flag.Var(&cfg.TrustedSubnet, "t", "Trusted subnet in CIDR")
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
	if len(cfg.Addr) == 0 {
		cfg.Addr = defaultAddr
	}
	if len(cfg.BaseURL) == 0 {
		cfg.BaseURL = defaultBaseURL
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
