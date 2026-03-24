package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/tini-yu/urlshrink/internal/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	// Драйвер для уже открытой *sql.DB
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("не удалось создать postgres-драйвер миграций: %w", err)
	}

	// Источник теперь из embed.FS
	sourceDriver, err := iofs.New(migrationsFS, "migrations") // "migrations" — имя папки внутри embed
	if err != nil {
		return fmt.Errorf("не удалось создать source driver из embed: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", // тип source
		sourceDriver,
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
