package handler

import (
	"context"
	"net/http"
	"time"
)

// Проверка подключения к БД
func (s *Shortener) PingDatabase(res http.ResponseWriter, req *http.Request) {

	if s.db == nil {
		http.Error(res, "База данных не запущена", http.StatusInternalServerError)
		return
	}

	// health-check
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		http.Error(res, "Неудалось подключиться к базе данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
