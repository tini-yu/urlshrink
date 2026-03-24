package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tini-yu/urlshrink/internal/logger"
)

// принимать в теле запроса JSON-объект {"url":"<some_url>"} и возвращать  {"result":"<short_url>"}.

type CreateShortURLRequest struct {
	URL string `json:"url"`
}

type CreateShortURLResponse struct {
	Result string `json:"result"`
}

func (s *Shortener) CreateShortURLJSON(res http.ResponseWriter, req *http.Request) {
	var input CreateShortURLRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		http.Error(res, "ошибка: некорректный JSON", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(input.URL)
	if originalURL == "" {
		http.Error(res, "ошибка: пустой URL", http.StatusBadRequest)
		return
	}

	// Используем общую логику из shortener
	shortURL, status, err := s.createOrGetShortURL(originalURL)
	if err != nil {
		logger.S.Errorw("Ошибка создания короткой ссылки (JSON): %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Для обоих статусов (201 и 409) возвращаем JSON в одинаковом формате
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)

	response := CreateShortURLResponse{Result: shortURL}
	if encodeErr := json.NewEncoder(res).Encode(response); encodeErr != nil {
		logger.S.Errorw("Ошибка кодирования JSON ответа: %v", encodeErr)
	}
}
