package main

import (
	"net/http"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg := config.Parse()
	shortener := handler.NewShortener(cfg)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Post("/", shortener.CreateShortURL)
		r.Get("/{id}", handler.GetFullURL)
	})

	http.ListenAndServe(cfg.HTTPAddr, r)
}
