package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/handler"
	"github.com/tini-yu/urlshrink/internal/logger"
	mware "github.com/tini-yu/urlshrink/internal/middleware"
	"github.com/tini-yu/urlshrink/migrations"
	"github.com/tini-yu/urlshrink/internal/storage"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Parse()

	// логгер zap
	logger.Init(false)
	defer logger.Sync()

	zap.RedirectStdLog(logger.L)

	// Подключение к БД:
	var db *sql.DB
	if cfg.DBPath != "" {
		var err error
		db, err = sql.Open("pgx", cfg.DBPath)
		if err != nil {
			log.Fatalf("Не удалось подключиться к базе данных: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			log.Fatalf("Не удалось подключиться к базе данных (Ping failed): %v\nDSN: %s", err, cfg.DBPath)
		}

		logger.L.Info("Успешно подключено к PostgreSQL")

		if err := migrations.RunMigrations(db); err != nil {
			logger.L.Fatal("Ошибка миграций: ", zap.Error(err))
		}
	}

	// Переключение между файловой и БД:
	var store storage.URLStorageInterface
	if cfg.DBPath != "" {
		store = storage.NewPostgresURLStorage(db)
		logger.L.Info("Используется хранилище: PostgreSQL")
	} else {
		var err error
		store, err = storage.NewFileURLStorage(cfg.URLFilePath)
		if err != nil {
			log.Fatalf("Не удалось инициализировать файловое хранилище: %v", err)
		}
		logger.L.Info("Используется хранилище: файл")
	}

	shortener := handler.NewShortener(store, cfg, db)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(mware.ZapLoggerMiddleware(logger.L))

	r.With(mware.GzipMiddleware).Route("/", func(r chi.Router) {
		r.Post("/", shortener.CreateShortURL)
		r.Get("/{id}", shortener.GetFullURL)
		r.Post("/api/shorten", shortener.CreateShortURLJSON)
		r.Post("/api/shorten/batch", shortener.ShortenBatch)
	})
	r.Get("/ping", shortener.PingDatabase)

	logger.L.Info("Сервер запущен", zap.String("address", cfg.HTTPAddr))
	err := http.ListenAndServe(cfg.HTTPAddr, r)
	if err != nil {
		logger.L.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}
