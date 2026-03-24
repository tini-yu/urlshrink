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
// При конфликте по original_url всегда возвращает ранее выданный short_url + 409 сатус.
func (s *Shortener) createOrGetShortURL(originalURL string) (shortURL string, status int, err error) {
	const maxRetries = 20

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortID := s.createShortID()

		fullShortURL, joinErr := url.JoinPath(s.cfg.BaseShortURL, shortID)
		if joinErr != nil {
			continue
		}

		// PostgreSQL
		if pgStore, ok := s.storage.(*storage.PostgresURLStorage); ok {
			finalShortID, isNew, dbErr := pgStore.GetOrCreateShortURL(shortID, originalURL)
			if dbErr != nil {
				return "", 0, fmt.Errorf("postgres GetOrCreateShortURL: %w", dbErr)
			}

			fullURL, _ := url.JoinPath(s.cfg.BaseShortURL, finalShortID)

			if isNew {
				return fullURL, http.StatusCreated, nil
			}

			// Конфликт - возвращаем уже существующий short_url
			return fullURL, http.StatusConflict, nil
		}

		// Файловое хранилище
		err = s.storage.SetIfNotExists(shortID, originalURL)
		if err == nil {
			return fullShortURL, http.StatusCreated, nil
		}
		if errors.Is(err, storage.ErrKeyAlreadyExists) || errors.Is(err, storage.ErrShortURLExists) {
			continue
		}
		return "", 0, err
	}

	return "", 0, fmt.Errorf("не удалось сгенерировать уникальный shortID после %d попыток", maxRetries)
}
