package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tini-yu/urlshrink/internal/storage"
)

func TestGetFullURL(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		id           string
		setupStorage func() // т.к. надо уже иметь данные в мапе
		wantStatus   int
		wantHeader   string // Location
		wantErr      string
	}{
		{
			name:   "Успешный редирект",
			method: http.MethodGet,
			id:     "550e8400-e29b-41d4-a716-446655440000",
			setupStorage: func() {
				storage.ShrunkURLs["550e8400-e29b-41d4-a716-446655440000"] = "http://some.website"
			},
			wantStatus: http.StatusTemporaryRedirect,
			wantHeader: "http://some.website",
			wantErr:    "",
		},
		{
			name:         "Неправильный {id}",
			method:       http.MethodGet,
			id:           "какой-то-неправильный-id",
			setupStorage: func() {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "неверный URL",
		},
		{
			name:         "Пустой {id}",
			method:       http.MethodGet,
			id:           "",
			setupStorage: func() {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "отсутсвует id",
		},
		{
			name:         "Hе GET метод",
			method:       http.MethodPost,
			id:           "/id",
			setupStorage: func() {},
			wantStatus:   http.StatusBadRequest,
			wantErr:      "Только GET запросы!",
		},
		{
			name:   "Существует id, но значение пустое",
			method: http.MethodGet,
			id:     "rfrjt-nj-bl",
			setupStorage: func() {
				storage.ShrunkURLs["rfrjt-nj-bl"] = ""
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "неверный URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage.ShrunkURLs = make(map[string]string)

			if test.setupStorage != nil {
				test.setupStorage()
			}

			req := httptest.NewRequest(test.method, "/", nil)
			newr := httptest.NewRecorder()

			getFullURLLogic(newr, req, test.id)

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
