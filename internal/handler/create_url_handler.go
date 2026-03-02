package handler

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"github.com/tini-yu/urlshrink/internal/storage"
)

const (
	charset         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCodeLength = 6
)

func (s *Shortener) createShortID() string {
		code := randomString(shortCodeLength)
		return code

}

func randomString(length int) string {
	sb := strings.Builder{}
	sb.Grow(length)

	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		sb.WriteByte(charset[n.Int64()])
	}
	return sb.String()
}

func (s *Shortener) CreateShortURL(res http.ResponseWriter, req *http.Request) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	originalURL := string(bodyBytes)
	originalURL = strings.TrimSpace(originalURL)
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

		//макс вызовов:
		if attempt == maxRetries-1 {
			log.Printf("Не удалось создать уникальный shortID после %d попыток", maxRetries)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// другая ошибка
		log.Printf("Ошибка сохранения (не коллизия): %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, convertedURL)
}
