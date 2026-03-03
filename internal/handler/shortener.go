package handler

import (
	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

type Shortener struct {
	cfg     config.Config
	storage *storage.FileURLStorage
}

func NewShortener(storage *storage.FileURLStorage, cfg config.Config) *Shortener {
	return &Shortener{
		storage: storage,
		cfg:   cfg,
	}
}
