package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tini-yu/urlshrink/internal/storage"
)

func CreateShortURL(res http.ResponseWriter, req *http.Request) {

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	originalURL := string(bodyBytes)
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		http.Error(res, "ошибка: пустой URL", http.StatusBadRequest)
		return
	}

	uuid := uuid.New().String()

	convertedURL := fmt.Sprintf("http://localhost:8080/%s", uuid)
	storage.ShrunkURLs[uuid] = originalURL

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, convertedURL)
}
