package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

func TestCreateShortURLJSON(t *testing.T) {
	type want struct {
		status     int
		bodyPrefix string // начало тела ответа (для успешного случая)
		errorMsg   string
		jsonResult bool
		mapUpdated bool
		checkShort func(*testing.T, string, *storage.FileURLStorage)
	}

	tests := []struct {
		name        string
		method      string
		body        string
		contentType string
		want        want
	}{
		{
			name:        "успешное создание (валидный JSON)",
			method:      http.MethodPost,
			body:        `{"url": "https://example.com/very/long/path"}`,
			contentType: "application/json",
			want: want{
				status:     http.StatusCreated,
				bodyPrefix: `"result":"`,
				jsonResult: true,
				mapUpdated: true,
				checkShort: func(t *testing.T, shortID string, st *storage.FileURLStorage) {
					orig, ok := st.GetOriginalURL(shortID)
					assert.True(t, ok, "должна быть запись в хранилище")
					assert.Equal(t, "https://example.com/very/long/path", orig)
				},
			},
		},
		{
			name:        "пустой url в JSON",
			method:      http.MethodPost,
			body:        `{"url": "   "}`,
			contentType: "application/json",
			want: want{
				status:     http.StatusBadRequest,
				errorMsg:   "ошибка: пустой URL",
				mapUpdated: false,
			},
		},
		{
			name:        "отсутствует поле url",
			method:      http.MethodPost,
			body:        `{"something": "else"}`,
			contentType: "application/json",
			want: want{
				status:     http.StatusBadRequest,
				errorMsg:   "ошибка: пустой URL",
				mapUpdated: false,
			},
		},
		{
			name:        "некорректный JSON",
			method:      http.MethodPost,
			body:        `{"url": "https://example.com" broken json`,
			contentType: "application/json",
			want: want{
				status:     http.StatusBadRequest,
				errorMsg:   "ошибка: некорректный JSON",
				mapUpdated: false,
			},
		},
		{
			name:        "неверный Content-Type (но text/plain всё равно парсится)",
			method:      http.MethodPost,
			body:        `{"url": "https://valid.url"}`,
			contentType: "text/plain",
			want: want{
				status:     http.StatusCreated,
				bodyPrefix: `"result":"`,
				jsonResult: true,
				mapUpdated: true,
			},
		},
		{
			name:        "не POST метод (Get)",
			method:      http.MethodGet,
			body:        "",
			contentType: "",
			want: want{
				status:     http.StatusMethodNotAllowed,
				mapUpdated: false,
			},
		},
		{
			name:        "пустое тело",
			method:      http.MethodPost,
			body:        "",
			contentType: "application/json",
			want: want{
				status:     http.StatusBadRequest,
				errorMsg:   "ошибка: некорректный JSON",
				mapUpdated: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBaseURL := "http://short.url"
			cfg := config.Config{BaseShortURL: testBaseURL}
			file := filepath.Join(t.TempDir(), "urls.json")
			st, _ := storage.NewFileURLStorage(file)
			h := NewShortener(st, cfg)

			r := chi.NewRouter()
			r.Post("/api/shorten", h.CreateShortURLJSON)

			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/api/shorten", body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want.status, w.Code, "статус ответа не совпал")

			respBody := w.Body.String()

			if tt.want.errorMsg != "" {
				assert.Contains(t, respBody, tt.want.errorMsg)
			}

			if tt.want.bodyPrefix != "" {
				assert.Contains(t, respBody, tt.want.bodyPrefix)
				assert.Contains(t, respBody, `"result":`)
				assert.True(t, strings.Contains(respBody, testBaseURL+"/"))
			}

			if tt.want.jsonResult {
				var got CreateShortURLResponse
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err, "ответ должен быть валидным JSON")
				assert.True(t, strings.HasPrefix(got.Result, testBaseURL+"/"))
				assert.True(t, len(got.Result) > len(testBaseURL)+1)
			}

			if tt.want.mapUpdated {
				assert.Equal(t, 1, st.Len(), "должна появиться ровно одна запись")
				keys := st.GetKeys()
				if len(keys) > 0 && tt.want.checkShort != nil {
					shortID := strings.TrimPrefix(keys[0], "/")
					tt.want.checkShort(t, shortID, st)
				}
			} else {
				assert.True(t, st.IsEmpty() || st.Len() == 0)
			}
		})
	}
}
