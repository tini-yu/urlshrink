package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Shortener) GetFullURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	s.getFullURLLogic(res, req, id)
}

func (s *Shortener) getFullURLLogic(res http.ResponseWriter, req *http.Request, id string) {

	originalURL, ok := s.storage.GetOriginalURL(id)
	if originalURL == "" || !ok {
		http.Error(res, "неверный URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
