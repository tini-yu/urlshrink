package config

import (
	"flag"
	"log"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr     string
	BaseShortURL string
	ServerPort   string
	URLFilePath  string
	DBPath       string
}

// Приоритет на переменную окружения, потом флаг, потом по умолчанию
func Parse() Config {
	var cfg Config

	flag.StringVar(&cfg.HTTPAddr, "a", "localhost:8080", "адрес запуска http сервера")
	flag.StringVar(&cfg.BaseShortURL, "b", "http://localhost:8080", "базовый адрес коротких ссылок")
	flag.StringVar(&cfg.ServerPort, "server-port", "", "порт сервера (перезапись -a)")
	flag.StringVar(&cfg.URLFilePath, "f", "urls_json", "путь до файла с URL")
	flag.StringVar(&cfg.DBPath, "d", "", "адрес базы данных")

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

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.HTTPAddr = envRunAddr
		log.Printf("Переопределен адрес сервера: %s", cfg.HTTPAddr)
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseShortURL = envBaseURL
		log.Printf("Переопределен базовый адрес коротких ссылок: %s", cfg.BaseShortURL)
	}
	if envURLFilePath := os.Getenv("FILE_STORAGE_PATH"); envURLFilePath != "" {
		cfg.URLFilePath = envURLFilePath
		log.Printf("Переопределен путь до файла с URL: %s", cfg.URLFilePath)
	}
	if envDBPath := os.Getenv("DATABASE_DSN"); envDBPath != "" {
		cfg.DBPath = envDBPath
		log.Printf("Адрес БД: %s", cfg.URLFilePath)
	}

	return cfg
}
