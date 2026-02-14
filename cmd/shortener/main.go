package main

import (
	"net/http"

	"github.com/tini-yu/urlshrink/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Post("/", handler.CreateShortURL)
		r.Get("/{id}", handler.GetFullURL)
	})

	http.ListenAndServe(":8080", r)
}
