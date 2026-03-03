package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tini-yu/urlshrink/internal/storage"
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

	const maxRetries = 20
	convertedURL := ""

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortID := s.createShortID()

		shortURL, err := url.JoinPath(s.cfg.BaseShortURL, shortID)
		if err != nil {
			log.Printf("Ошибка JoinPath (attempt %d): %v", attempt+1, err)
			continue
		}

		// ошибка при сете
		err = s.storage.SetIfNotExists(shortID, originalURL)
		if err == nil {
			convertedURL = shortURL
			break
		}

		// если коллизия - повторяем цикл
		if errors.Is(err, storage.ErrKeyAlreadyExists) {
			log.Printf("Коллизия shortID %s, попытка %d", shortID, attempt+1)
			continue
		}

		// другая ошибка
		log.Printf("Ошибка сохранения (не коллизия): %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// макс вызовов:
	if convertedURL == "" {
		log.Printf("Не удалось создать уникальный shortID после %d попыток", maxRetries)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	response := CreateShortURLResponse{Result: convertedURL}
	_ = json.NewEncoder(res).Encode(response)
}
