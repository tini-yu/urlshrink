package main

import (
	"net/http"

	"github.com/tini-yu/urlshrink/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.CreateShortURL)
	mux.HandleFunc("/{id}", handler.GetFullURL)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}
