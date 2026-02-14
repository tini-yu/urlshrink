package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// Мапа уже сконвертированных url:
var URLs = make(map[string]string)

func getFullURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Только GET запросы!", http.StatusBadRequest)
		return
	}
	id := req.PathValue("id") // т.к. смотрит /{id}
	if id == "" {
		http.Error(res, "отсутсвует id", http.StatusBadRequest)
		return
	}

	originalURL, exists := URLs[id]
	if originalURL == "" || !exists {
		http.Error(res, "неверный URL", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func createShortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Только POST запросы!", http.StatusBadRequest)
		return
	}

	// Читаем тело запроса (text/plain)
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}

	originalURL := string(bodyBytes)
	if originalURL == "" {
		http.Error(res, "ошибка: пустой URL", http.StatusBadRequest)
		return
	}

	// uuid для уникального нового url
	uuid := uuid.New().String()

	convertedURL := fmt.Sprintf("http://localhost:8080/%s", uuid)
	URLs[uuid] = originalURL

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprintf(res, convertedURL)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", createShortURL)
	mux.HandleFunc("/{id}", getFullURL)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}
