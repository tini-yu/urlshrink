package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type compressWriter struct {
	io.Writer
	http.ResponseWriter
	sync.Once
}

func (w *compressWriter) Write(b []byte) (int, error) {
	w.Once.Do(func() {
		//ставим заголовок только при первой записи
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
	})
	return w.Writer.Write(b)
}

func (w *compressWriter) Flush() error {
	if fl, ok := w.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
		return nil
	}
	return nil
}

type compressReader struct {
	io.ReadCloser
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	gz := gzip.NewWriter(w)
	return &compressWriter{
		Writer:         gz,
		ResponseWriter: w,
	}
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{ReadCloser: gz}, nil
}

func (r compressReader) Close() error {
	return r.ReadCloser.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w // оригинальный writer

		// сжатие ответа
		accept := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(accept, "gzip")

		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Writer.(*gzip.Writer).Close()
		}

		// распаковка входящего тела
		contentEnc := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEnc, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "cannot decode gzip body", http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
