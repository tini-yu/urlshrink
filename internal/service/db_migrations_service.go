package service

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/tini-yu/urlshrink/internal/logger"
)

func RunMigrations(db *sql.DB) error {
	// Драйвер для уже открытой *sql.DB
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("не удалось создать postgres-драйвер миграций: %w", err)
	}
	const migrationsPath = "./../../migrations"
	// Источник миграций — файлы
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать миграции: %w", err)
	}
	
	// Применяем все миграции вверх
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}
	
	logger.L.Info("Миграции успешно применены")
	return nil
}
