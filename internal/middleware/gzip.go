package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// распаковка если gzip прислали
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzr, err := newCompReader(r.Body)
			if err != nil {
				http.Error(w, "cannot decode gzip request body", http.StatusBadRequest)
				return
			}
			r.Body = gzr
			defer gzr.Close()
		}

		// сжатие ТОЛЬКО при поддержке gzip
		var zw http.ResponseWriter = w
		var cw *compWriter

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw = newCompWriter(w)
			zw = cw
			defer cw.Close()
		}

		next.ServeHTTP(zw, r)
	})
}

// writer для сжатия

type compWriter struct {
	w  http.ResponseWriter
	zw io.WriteCloser // gzip.Writer
}

func newCompWriter(w http.ResponseWriter) *compWriter {
	return &compWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compWriter) Write(b []byte) (int, error) {
	return c.zw.Write(b)
}

func (c *compWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compWriter) Close() error {
	return c.zw.Close()
}

// распаковываем

type compReader struct {
	r  io.ReadCloser // оригинальный body
	zr io.ReadCloser // gzip reader
}

func newCompReader(r io.ReadCloser) (*compReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compReader) Read(b []byte) (int, error) {
	return c.zr.Read(b)
}

func (c *compReader) Close() error {
	// закрываем gzip-ридер ( не оригинальный body)
	err := c.zr.Close()
	return err
}