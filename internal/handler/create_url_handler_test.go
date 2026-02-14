package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
		uuidCheck     func(t *testing.T, createdUUID string)
	}{
		{
			name:          "Успешное сокращение",
			method:        http.MethodPost,
			body:          "http://some.website",
			wantStatus:    http.StatusCreated,
			wantBodyPref:  "http://localhost:8080/",
			wantMapUpdate: true,
			uuidCheck: func(t *testing.T, createdUUID string) {
				original, exists := storage.ShrunkURLs[createdUUID]
				assert.True(t, exists)
				assert.Equal(t, "http://some.website", original)
			},
		},
		{
			name:          "Не POST метод",
			method:        http.MethodGet,
			body:          "",
			wantStatus:    http.StatusBadRequest,
			wantErr:       "Только POST запросы!",
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
			storage.ShrunkURLs = make(map[string]string)

			var body io.Reader
			if test.body != "" && test.method == http.MethodPost {
				body = strings.NewReader(test.body)
			}

			req := httptest.NewRequest(test.method, "/api/to/shorten", body)
			req.Header.Set("Content-Type", "text/plain")

			newr := httptest.NewRecorder()

			CreateShortURL(newr, req)

			assert.Equal(t, test.wantStatus, newr.Code)

			responseBody := newr.Body.String()
			if test.wantErr != "" {
				assert.Contains(t, responseBody, test.wantErr)
			}
			if test.wantBodyPref != "" {
				assert.Contains(t, responseBody, test.wantBodyPref)
				assert.True(t, strings.HasPrefix(responseBody, "http://localhost:8080/"))
			}
			assert.Equal(t, "text/plain; charset=utf-8", newr.Header().Get("Content-Type"))

			if test.wantMapUpdate {
				assert.NotEmpty(t, storage.ShrunkURLs)
				assert.Len(t, storage.ShrunkURLs, 1)

				// uuid проверка
				for key := range storage.ShrunkURLs {
					_, err := uuid.Parse(key)
					assert.NoError(t, err, "ключ должен быть валидным UUID")

					if test.uuidCheck != nil {
						test.uuidCheck(t, key)
					}
				}
			} else {
				assert.Empty(t, storage.ShrunkURLs)
			}
		})
	}
}
