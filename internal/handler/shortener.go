package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

type Shortener struct {
	cfg     config.Config
	storage storage.URLStorageInterface
	db      *sql.DB
}

func NewShortener(storage storage.URLStorageInterface, cfg config.Config, db *sql.DB) *Shortener {
	return &Shortener{
		storage: storage,
		cfg:     cfg,
		db:      db,
	}
}

// createOrGetShortURL - общая логика для хендлеров create_url | create_json_url
// При конфликте по original_url всегда возвращает ранее выданный short_url + 409 статус.
func (s *Shortener) createOrGetShortURL(originalURL string) (shortURL string, status int, err error) {
	const maxRetries = 20

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortID := s.createShortID()

		// Единый вызов через интерфейс
		finalShortID, isNew, storeErr := s.storage.GetOrCreateShortURL(shortID, originalURL)
		if storeErr != nil {
			if errors.Is(storeErr, storage.ErrKeyAlreadyExists) {
				continue // коллизия по shortID - пробуем ещё раз
			}
			return "", 0, fmt.Errorf("storage CreateShortURL failed: %w", storeErr)
		}

		finalFullURL, _ := url.JoinPath(s.cfg.BaseShortURL, finalShortID)

		if isNew {
			return finalFullURL, http.StatusCreated, nil
		}

		// Конфликт по original_url (только Postgres)
		return finalFullURL, http.StatusConflict, nil
	}

	return "", 0, fmt.Errorf("не удалось сгенерировать уникальный shortID после %d попыток", maxRetries)
}
