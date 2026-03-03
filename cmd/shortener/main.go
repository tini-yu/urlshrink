package main

import (
	"log"
	"net/http"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/handler"
	mware "github.com/tini-yu/urlshrink/internal/middleware"
	"github.com/tini-yu/urlshrink/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Parse()

	// логгер zap
	zapLog, err := zap.NewDevelopment() // потом NewProduction() поменять
	if err != nil {
		log.Fatalf("Не получилось инициализировать zap: %v", err)
	}
	defer zapLog.Sync()

	zap.RedirectStdLog(zapLog)

	storage, err := storage.NewFileURLStorage(cfg.URLFilePath)
	if err != nil {
		log.Fatalf("Не удалось инициализировать файловое хранилище: %v", err)
	}
	shortener := handler.NewShortener(storage, cfg)

	r := chi.NewRouter()
	r.Use(mware.GzipMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(mware.ZapLoggerMiddleware(zapLog))

	r.Route("/", func(r chi.Router) {
		r.Post("/", shortener.CreateShortURL)
		r.Get("/{id}", shortener.GetFullURL)
		r.Post("/api/shorten", shortener.CreateShortURLJSON)
	})

	zapLog.Info("Сервер запущен", zap.String("address", cfg.HTTPAddr))
	err = http.ListenAndServe(cfg.HTTPAddr, r)
	if err != nil {
		zapLog.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}
