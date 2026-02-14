package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tini-yu/urlshrink/internal/storage"
)

func GetFullURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	getFullURLLogic(res, req, id)
}

func getFullURLLogic(res http.ResponseWriter, req *http.Request, id string) {
	if id == "" {
		http.Error(res, "отсутсвует id", http.StatusBadRequest)
		return
	}

	originalURL, exists := storage.ShrunkURLs[id]
	if originalURL == "" || !exists {
		http.Error(res, "неверный URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
