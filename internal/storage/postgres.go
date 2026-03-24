package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tini-yu/urlshrink/internal/logger"
	"go.uber.org/zap"
)

var ErrShortURLExists = errors.New("короткая ссылка уже существует")

type PostgresURLStorage struct {
	db *sql.DB
}

func NewPostgresURLStorage(db *sql.DB) *PostgresURLStorage {
	return &PostgresURLStorage{db: db}
}

func (s *PostgresURLStorage) GetOriginalURL(shortID string) (string, bool) {
	var original string
	err := s.db.QueryRowContext(context.Background(),
		`SELECT original_url FROM urls WHERE short_url = $1`,
		shortID,
	).Scan(&original)

	if err == sql.ErrNoRows {
		logger.L.Error("Не найдена ни одна строка с оригинальным url", zap.Error(err))
		return "", false
	}
	if err != nil {
		logger.L.Error("Ошибка поиска оригинальной url", zap.Error(err))
		return "", false
	}
	return original, true
}

func (s *PostgresURLStorage) CheckShortURL(shortID string) bool {
	var cnt int
	_ = s.db.QueryRowContext(context.Background(),
		`SELECT 1 FROM urls WHERE short_url = $1 LIMIT 1`,
		shortID,
	).Scan(&cnt)
	return cnt == 1
}

func (s *PostgresURLStorage) SetIfNotExists(shortID, originalURL string) error {
	ctx := context.Background()

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO urls (short_url, original_url)
		 VALUES ($1, $2)
		 ON CONFLICT (short_url) DO NOTHING`,
		shortID, originalURL,
	)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		// значит уже существовал
		return ErrShortURLExists
	}

	return nil
}

// Заглушки для остальных методов интерфейса (можно реализовать позже)
func (s *PostgresURLStorage) SetURL(shortID, originalURL string) {
	// не используется в твоём текущем коде
}

func (s *PostgresURLStorage) Len() int {
	return -1
}

func (s *PostgresURLStorage) IsEmpty() bool {
	return false
}

func (s *PostgresURLStorage) GetKeys() []string {
	return nil
}

// GetOrCreateShortURL - вставляет или возвращает существующий short_url по original_url
// Возвращает существующий shortID, isNew (true = вставили новую запись), error
func (s *PostgresURLStorage) GetOrCreateShortURL(proposedShortID, originalURL string) (string, bool, error) {
	ctx := context.Background()

	var finalShortID string
	var isNew bool

	err := s.db.QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO urls (short_url, original_url)
			VALUES ($1, $2)
			ON CONFLICT (original_url) DO NOTHING
			RETURNING short_url
		)
		SELECT short_url, true AS is_new 
		FROM inserted
		
		UNION ALL
		
		SELECT short_url, false AS is_new 
		FROM urls 
		WHERE original_url = $2 
		  AND NOT EXISTS (SELECT 1 FROM inserted)
		LIMIT 1
	`, proposedShortID, originalURL).Scan(&finalShortID, &isNew)

	if err != nil {
		return "", false, fmt.Errorf("insert or conflict failed: %w", err)
	}

	return finalShortID, isNew, nil
}
