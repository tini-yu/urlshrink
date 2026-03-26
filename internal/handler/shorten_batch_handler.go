package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/tini-yu/urlshrink/internal/logger"
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

	createItems := make([]storage.BatchCreateItem, 0, len(batch))
	const maxRetries = 20

	for _, item := range batch {
		originalURL := strings.TrimSpace(item.OriginalURL)
		if originalURL == "" {
			// пропускаем, если url пустой
			logger.S.Infow("Пропущен пустой URL для correlation_id: %s", item.CorrelationID)
			continue
		}

		// генерим новый id
		var shortID string
		for attempt := 0; attempt < maxRetries; attempt++ {
			candidate := s.createShortID()
			if !s.storage.CheckShortURL(candidate) { // чек на всякий
				shortID = candidate
				break
			}
		}
		if shortID == "" {
			logger.S.Errorw("Не удалось создать уникальный shortID после %d попыток для correlation_id: %s", item.CorrelationID)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		createItems = append(createItems, storage.BatchCreateItem{
			CorrelationID: item.CorrelationID,
			ShortID:       shortID,
			OriginalURL:   originalURL,
		})
	}

	if len(createItems) == 0 {
		http.Error(res, "все URL в батче были пустыми", http.StatusBadRequest)
		return
	}

	// одна запись для батча в postgres
	results, err := s.storage.BatchCreateShortURLs(createItems)
	if err != nil {
		logger.S.Errorw("Ошибка batch create: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := make([]BatchShortenResponseItem, 0, len(results))
	for _, resItem := range results {
		fullShortURL, _ := url.JoinPath(s.cfg.BaseShortURL, resItem.ShortID)

		response = append(response, BatchShortenResponseItem{
			CorrelationID: resItem.CorrelationID,
			ShortURL:      fullShortURL,
		})
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(response); err != nil {
		logger.S.Errorw("Ошибка кодирования ответа batch: %v", err)
	}

}