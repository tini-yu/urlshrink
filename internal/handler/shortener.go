package handler

import (
	"database/sql"

	"github.com/tini-yu/urlshrink/internal/config"
	"github.com/tini-yu/urlshrink/internal/storage"
)

type Shortener struct {
	cfg     config.Config
	storage storage.URLStorageInterface
	db      *sql.DB
}

func NewShortener(storage storage.URLStorageInterface, cfg config.Config, db *sql.DB) *Shortener {
	return &Shortener{
		storage: storage,
		cfg:     cfg,
		db:      db,
	}
}
