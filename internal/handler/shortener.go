package handler

import "github.com/tini-yu/urlshrink/internal/config"

type Shortener struct {
	cfg config.Config
}

func NewShortener(cfg config.Config) *Shortener {
	return &Shortener{cfg: cfg}
}