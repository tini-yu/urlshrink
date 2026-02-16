package main

import (
	"log"
	"net/http"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/handler"
	"github.com/tini-yu/urlshrink/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Parse()
	storage := storage.NewURLStorage()
	shortener := handler.NewShortener(storage, cfg)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Post("/", shortener.CreateShortURL)
		r.Get("/{id}", shortener.GetFullURL)
	})

	log.Printf("Сервер запущен по адресу = %s", cfg.HTTPAddr)
	err := http.ListenAndServe(cfg.HTTPAddr, r)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
