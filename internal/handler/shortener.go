package handler

import (
	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

type Shortener struct {
	cfg     config.Config
	storage *storage.URLStorage
}

func NewShortener(storage *storage.URLStorage, cfg config.Config) *Shortener {
	return &Shortener{
		storage: storage,
		cfg:   cfg,
	}
}
