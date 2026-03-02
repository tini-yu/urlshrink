package storage

import "errors"

type URLStorage struct {
	urls map[string]string //[короткая]: оригинал
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urls: make(map[string]string),
	}
}

func (s *URLStorage) GetOriginalURL(shortID string) (string, bool) {
	original, ok := s.urls[shortID]
	return original, ok
}

func (s *URLStorage) SetURL(shortID, originalURL string) {
	s.urls[shortID] = originalURL
}

func (s *URLStorage) CheckShortURL(shortID string) bool {
	_, ok := s.urls[shortID]
	return ok
}

func (s *URLStorage) Len() int {
	return len(s.urls)
}

func (s *URLStorage) IsEmpty() bool {
	return s.Len() == 0
}

func (s *URLStorage) GetKeys() []string {
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys
}

var (
	ErrKeyAlreadyExists = errors.New("short id already exists")
)

func (s *URLStorage) SetIfNotExists(shortID, originalURL string) error {

	if _, ok := s.urls[shortID]; ok {
		return ErrKeyAlreadyExists
	}

	s.urls[shortID] = originalURL
	return nil
}