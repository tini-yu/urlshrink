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
	// if id == "" {
	// 	http.Error(res, "отсутсвует id", http.StatusBadRequest)
	// 	return
	// } //Вроде никогда не сработает, т.к. пустой id = Page Not Found

	originalURL, ok := s.storage.GetOriginalURL(id)
	if originalURL == "" || !ok {
		http.Error(res, "неверный URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
