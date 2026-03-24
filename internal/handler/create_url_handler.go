package handler

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/tini-yu/urlshrink/internal/logger"
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
    bodyBytes, _ := io.ReadAll(req.Body)
    defer req.Body.Close()

    originalURL := strings.TrimSpace(string(bodyBytes))
    if originalURL == "" {
        http.Error(res, "ошибка: пустой URL", http.StatusBadRequest)
        return
    }

    shortURL, status, err := s.createOrGetShortURL(originalURL)
    if err != nil {
        logger.S.Errorw("Ошибка создания URL: %v", err)
        http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }

    res.Header().Set("Content-Type", "text/plain; charset=utf-8")
    res.WriteHeader(status)
    fmt.Fprint(res, shortURL)
}