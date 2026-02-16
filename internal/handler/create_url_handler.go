package handler

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"

)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const shortCodeLength = 6

func (s *Shortener) createShortID() string {
    for {
        code := randomString(shortCodeLength)
        if _, exists := s.storage.GetOriginalURL(code); !exists {
            return code
        }
    }
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

	shortID := s.createShortID()

	convertedURL, err := url.JoinPath(s.cfg.BaseShortURL, shortID)
	if err != nil {
		log.Printf("Ошибка JoinPath: %v", err)
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	s.storage.SetURL(shortID, originalURL)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, convertedURL)
}
