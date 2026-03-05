package storage

import (
	"encoding/json"
	"errors"
	"os"
)

// стракт одного вхождения в файле
type URLRecord struct {
	// UUID        string `json:"uuid"`         // по заданию надо
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileURLStorage struct {
	filePath string
	urls     map[string]string          // [короткая]: оригинал
	records  []URLRecord                
}

type URLStorageInterface interface {
	GetOriginalURL(shortID string) (string, bool)
	SetURL(shortID, originalURL string)
    CheckShortURL(shortID string) bool           
    Len() int                                     
    IsEmpty() bool                              
    GetKeys() []string                                       
    SetIfNotExists(shortID, originalURL string) error
}

func NewFileURLStorage(filePath string) (*FileURLStorage, error) {
	s := &FileURLStorage{
		filePath: filePath,
		urls:     make(map[string]string),
		records:  []URLRecord{},
	}

	if err := s.loadFromFile(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return s, nil
}

func (s *FileURLStorage) loadFromFile() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var records []URLRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	s.records = records
	s.urls = make(map[string]string, len(records))

	for _, rec := range records {
		if rec.ShortURL != "" {
			s.urls[rec.ShortURL] = rec.OriginalURL
		}
	}

	return nil
}

// Перезаписываем весь файл
func (s *FileURLStorage) saveToFile() error {
	data, err := json.MarshalIndent(s.records, "", "  ")

	if err != nil {
		return err
	}

	// Используем временный файл → атомарная замена (лучше защищает от повреждения)
	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filePath)
}

func (s *FileURLStorage) GetOriginalURL(shortID string) (string, bool) {
	original, ok := s.urls[shortID]
	return original, ok
}

func (s *FileURLStorage) SetURL(shortID, originalURL string) {
	s.urls[shortID] = originalURL
}

func (s *FileURLStorage) CheckShortURL(shortID string) bool {
	_, ok := s.urls[shortID]
	return ok
}

func (s *FileURLStorage) Len() int {
	return len(s.urls)
}

func (s *FileURLStorage) IsEmpty() bool {
	return s.Len() == 0
}

func (s *FileURLStorage) GetKeys() []string {
	keys := make([]string, 0, len(s.urls))
	for k := range s.urls {
		keys = append(keys, k)
	}
	return keys
}

var (
	ErrKeyAlreadyExists = errors.New("short id already exists")
)

// Основной метод добавления — с проверкой уникальности и сохранением в файл
func (s *FileURLStorage) SetIfNotExists(shortID, originalURL string) error {

	if _, exists := s.urls[shortID]; exists {
		return ErrKeyAlreadyExists
	}

	// Добавляем в карту для быстрого поиска
	s.urls[shortID] = originalURL

	record := URLRecord{
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	s.records = append(s.records, record)

	return s.saveToFile()
}