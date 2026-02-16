package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

func TestGetFullURL(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		setupStorage func(storage *storage.URLStorage)
		wantStatus   int
		wantHeader   string // Location
		wantErr      string
	}{
		{
			name:   "Успешный редирект",
			method: http.MethodGet,
			path:   "/38fg338yf",
			setupStorage: func(storage *storage.URLStorage) {
				storage.SetURL("38fg338yf", "http://some.website")
			},
			wantStatus: http.StatusTemporaryRedirect,
			wantHeader: "http://some.website",
			wantErr:    "",
		},
		{
			name:         "Неправильный {id}",
			method:       http.MethodGet,
			path:         "/какой-то-неправильный-id",
			setupStorage: func(storage *storage.URLStorage) {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "неверный URL",
		},
		{
			name:         "Пустой {id}",
			method:       http.MethodGet,
			path:         "/",
			setupStorage: func(storage *storage.URLStorage) {},
			wantStatus:   http.StatusNotFound,
		},
		{
			name:         "Hе GET метод",
			method:       http.MethodPost,
			path:         "/id",
			setupStorage: func(storage *storage.URLStorage) {},
			wantStatus:   http.StatusMethodNotAllowed,
			wantErr:      "",
		},
		{
			name:   "Существует id, но значение пустое",
			method: http.MethodGet,
			path:   "/4bfsFae",
			setupStorage: func(storage *storage.URLStorage) {
				storage.SetURL("4bfsFae", "")
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "неверный URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := storage.NewURLStorage()

			if test.setupStorage != nil {
				test.setupStorage(storage)
			}

			h := NewShortener(storage, config.Config{}) // пустой конфиг сойдет
			r := chi.NewRouter()
			r.Get("/{id}", h.GetFullURL)

			req := httptest.NewRequest(test.method, test.path, nil)
			newr := httptest.NewRecorder()

			r.ServeHTTP(newr, req)

			assert.Equal(t, test.wantStatus, newr.Code, "код статуса не совпадает")

			body := newr.Body.String()
			if test.wantErr != "" {
				assert.Contains(t, body, test.wantErr, "тело ответа не содержит ожидаемую ошибку")
			}

			if test.wantHeader != "" {
				assert.Equal(t,
					test.wantHeader,
					newr.Header().Get("Location"),
					"Location заголовок неправильный",
				)
			} else {
				assert.Empty(t, newr.Header().Get("Location"), "Location не должен быть установлен")
			}
		})
	}
}
