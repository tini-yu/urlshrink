package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

func TestCreateShortURL(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		body          string
		wantStatus    int
		wantErr       string
		wantBodyPref  string
		wantMapUpdate bool
		//Проверка, что обе ссылки верные:
		shortIDCheck     func(t *testing.T, createdShortID string, store *storage.URLStorage)
	}{
		{
			name:          "Успешное сокращение",
			method:        http.MethodPost,
			body:          "http://some.website",
			wantStatus:    http.StatusCreated,
			wantBodyPref:  "http://localhost:9090",
			wantMapUpdate: true,
			shortIDCheck: func(t *testing.T, createdShortID string, store *storage.URLStorage) {
				original, ok := store.GetOriginalURL(createdShortID)
				assert.True(t, ok)
				assert.Equal(t, "http://some.website", original)
			},
		},
		{
			name:          "Не POST метод",
			method:        http.MethodGet,
			body:          "",
			wantStatus:    http.StatusMethodNotAllowed,
			wantErr:       "",
			wantMapUpdate: false,
		},
		{
			name:          "пустое тело запроса",
			method:        http.MethodPost,
			body:          "",
			wantStatus:    http.StatusBadRequest,
			wantErr:       "ошибка: пустой URL",
			wantMapUpdate: false,
		},
		{
			name:          "Тело из пробелов",
			method:        http.MethodPost,
			body:          "    \t\n        ",
			wantStatus:    http.StatusBadRequest,
			wantErr:       "ошибка: пустой URL",
			wantMapUpdate: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var body io.Reader
			if test.body != "" && test.method == http.MethodPost {
				body = strings.NewReader(test.body)
			}

			testBaseURL := "http://localhost:9090"
			cfg := config.Config{
				BaseShortURL: testBaseURL,
			}

			storage := storage.NewURLStorage()
			h := NewShortener(storage, cfg)

			r := chi.NewRouter()
			r.Post("/", h.CreateShortURL)

			req := httptest.NewRequest(test.method, "/", body)
			req.Header.Set("Content-Type", "text/plain")

			newr := httptest.NewRecorder()

			r.ServeHTTP(newr, req)

			assert.Equal(t, test.wantStatus, newr.Code)
			// При неверном методе заголовка нету
			if test.wantStatus == http.StatusCreated {
				assert.Equal(t, "text/plain; charset=utf-8", newr.Header().Get("Content-Type"))
			}

			responseBody := newr.Body.String()
			if test.wantErr != "" {
				assert.Contains(t, responseBody, test.wantErr)
			}
			if test.wantBodyPref != "" {
				assert.Contains(t, responseBody, test.wantBodyPref)
				assert.True(t, strings.HasPrefix(responseBody, "http://localhost:9090/"))
			}

			if test.wantMapUpdate {
				assert.False(t, storage.IsEmpty())
				assert.Equal(t, 1, storage.Len()) //одна запись

				key := storage.GetKeys()[0]
				if test.shortIDCheck != nil {
					test.shortIDCheck(t, key, storage)
				}

			} else {
				assert.True(t, storage.IsEmpty())
			}
		})
	}
}
