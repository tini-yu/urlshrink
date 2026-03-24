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

// на вход
type BatchShortenRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// на выход
type BatchShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (s *Shortener) ShortenBatch(res http.ResponseWriter, req *http.Request) {
	var batch []BatchShortenRequestItem
	if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
		http.Error(res, "ошибка: некорректный JSON", http.StatusBadRequest)
		return
	}

	if len(batch) == 0 {
		http.Error(res, "ошибка: пустой массив", http.StatusBadRequest)
		return
	}

	response := make([]BatchShortenResponseItem, 0, len(batch))
	const maxRetries = 20

	for _, item := range batch {
		originalURL := strings.TrimSpace(item.OriginalURL)
		if originalURL == "" {
			// пропускаем, если url пустой
			log.Printf("Пропущен пустой URL для correlation_id: %s", item.CorrelationID)
			continue
		}

		convertedURL := ""

		for attempt := 0; attempt < maxRetries; attempt++ {
			shortID := s.createShortID()

			shortURL, err := url.JoinPath(s.cfg.BaseShortURL, shortID)
			if err != nil {
				log.Printf("Ошибка JoinPath (attempt %d): %v", attempt+1, err)
				continue
			}

			err = s.storage.SetIfNotExists(shortID, originalURL)
			if err == nil {
				convertedURL = shortURL
				break
			}

			if errors.Is(err, storage.ErrKeyAlreadyExists) {
				log.Printf("Коллизия shortID %s, попытка %d", shortID, attempt+1)
				continue
			}

			// любая другая ошибка хранения
			log.Printf("Ошибка сохранения (не коллизия): %v", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if convertedURL == "" {
			log.Printf("Не удалось создать уникальный shortID после %d попыток для correlation_id: %s",
				maxRetries, item.CorrelationID)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response = append(response, BatchShortenResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      convertedURL,
		})
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(response); err != nil {
		log.Printf("Ошибка кодирования ответа batch: %v", err)
	}
}