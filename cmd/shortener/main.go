package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/handler"
	mware "github.com/tini-yu/urlshrink/internal/middleware"
	"github.com/tini-yu/urlshrink/internal/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	// Подключение к БД:
	var db *sql.DB
	if cfg.DBPath != "" {

		db, err = sql.Open("pgx", cfg.DBPath)
		if err != nil {
			log.Fatalf("Не удалось подключиться к базе данных: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			log.Fatalf("Не удалось подключиться к базе данных (Ping failed): %v\nDSN: %s", err, cfg.DBPath)
		}

		zapLog.Info("Успешно подключено к PostgreSQL")
	}

	storage, err := storage.NewFileURLStorage(cfg.URLFilePath)
	if err != nil {
		log.Fatalf("Не удалось инициализировать файловое хранилище: %v", err)
	}
	shortener := handler.NewShortener(storage, cfg, db)

	r := chi.NewRouter()
	// r.Use(mware.GzipMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(mware.ZapLoggerMiddleware(zapLog))

	r.With(mware.GzipMiddleware).Route("/", func(r chi.Router) {
		r.Post("/", shortener.CreateShortURL)
		r.Get("/{id}", shortener.GetFullURL)
		r.Post("/api/shorten", shortener.CreateShortURLJSON)
	})
	r.Get("/ping", shortener.PingDatabase)

	zapLog.Info("Сервер запущен", zap.String("address", cfg.HTTPAddr))
	err = http.ListenAndServe(cfg.HTTPAddr, r)
	if err != nil {
		zapLog.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}
