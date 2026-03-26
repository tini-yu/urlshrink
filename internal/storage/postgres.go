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

// Заглушки для остальных методов интерфейса
func (s *PostgresURLStorage) SetURL(shortID, originalURL string) {
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

type BatchCreateItem struct {
	CorrelationID string
	ShortID       string
	OriginalURL   string
}

type BatchCreateResult struct {
	CorrelationID string
	ShortID       string // финальный short_id (может отличаться от предложенного)
	IsNew         bool   // true = вставили новую запись
}

func (s *PostgresURLStorage) BatchCreateShortURLs(items []BatchCreateItem) ([]BatchCreateResult, error) {
	if len(items) == 0 {
		return nil, nil
	}

	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback() // безопасно, если коммит прошёл — игнорируется

	// Подготавливаем данные для UNNEST
	corrIDs := make([]string, len(items))
	shortIDs := make([]string, len(items))
	origURLs := make([]string, len(items))

	for i, item := range items {
		corrIDs[i] = item.CorrelationID
		shortIDs[i] = item.ShortID
		origURLs[i] = item.OriginalURL
	}

	// Один запрос: вставляем всё + возвращаем финальные short_id и признак is_new
	rows, err := tx.QueryContext(ctx, `
		WITH input_data AS (
			SELECT *
			FROM unnest($1::text[], $2::text[], $3::text[])
			AS t(correlation_id, proposed_short_id, original_url)
		),
		inserted AS (
			INSERT INTO urls (short_url, original_url)
			SELECT proposed_short_id, original_url
			FROM input_data
			ON CONFLICT (original_url) DO NOTHING
			RETURNING short_url, original_url
		)
		SELECT 
			id.correlation_id,
			COALESCE(ins.short_url, existing.short_url) AS final_short_id,
			(ins.short_url IS NOT NULL) AS is_new
		FROM input_data id
		LEFT JOIN inserted ins 
			ON ins.original_url = id.original_url
		LEFT JOIN urls existing 
			ON existing.original_url = id.original_url 
		   AND ins.short_url IS NULL
	`, corrIDs, shortIDs, origURLs)
	if err != nil {
		return nil, fmt.Errorf("не удалось insert батч: %w", err)
	}
	defer rows.Close()

	var results []BatchCreateResult
	for rows.Next() {
		var r BatchCreateResult
		if err := rows.Scan(&r.CorrelationID, &r.ShortID, &r.IsNew); err != nil {
			return nil, fmt.Errorf("scan строки не удался: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return results, nil
}
