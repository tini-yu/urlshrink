package config

import (
	"flag"
	"log"
	"strings"
)

type Config struct {
	HTTPAddr     string
	BaseShortURL string
	ServerPort   string
}

func Parse() Config {
	var cfg Config

	flag.StringVar(&cfg.HTTPAddr, "a", "localhost:8080", "адрес запуска http сервера")
	flag.StringVar(&cfg.BaseShortURL, "b", "http://localhost:8080", "базовый адрес коротких ссылок")
	flag.StringVar(&cfg.ServerPort, "server-port", "", "порт сервера (перезапись -a)")

	flag.Parse()

	if cfg.ServerPort != "" {
		if strings.Contains(cfg.HTTPAddr, ":") {
			host := strings.SplitN(cfg.HTTPAddr, ":", 2)[0]
			cfg.HTTPAddr = host + ":" + cfg.ServerPort
		} else {
			cfg.HTTPAddr = ":" + cfg.ServerPort
		}

		log.Printf("Переопределен порт: %s", cfg.ServerPort)
	}

	return cfg
}
