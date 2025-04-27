package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"log/slog"
)

const (
	// migrationDir is the migrations' directory name
	migrationDir = "migrations"

	// version is the current database migration version
	version = 1
)

// migrationFiles is populated when building the binary
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

func migrateDB(
	ctx context.Context,
	log *slog.Logger,
	dbName string,
) error {
	log.InfoContext(ctx, "running Sqlite DB migration")

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}

	client, err := buildMigrationClient(db)
	if err != nil {
		return fmt.Errorf("building migration client: %w", err)
	}

	defer func() { _, _ = client.Close() }()

	err = client.Migrate(version)
	if err != nil && errors.Is(err, migrate.ErrNoChange) {
		log.InfoContext(ctx, "Sqlite migration is already up-to-date", slog.Int("version", version))
		return nil
	}

	if err != nil {
		return fmt.Errorf("migrating DB: %w", err)
	}

	log.InfoContext(ctx, "Sqlite DB migration complete", slog.Int("version", version))

	return nil
}

// buildMigrationClient creates a new migrate instance.
// source and target connection are required to be closed it in the calling function
func buildMigrationClient(db *sql.DB) (*migrate.Migrate, error) {
	// Read the source for the migrations.
	// Our source is the SQL files in the migrations folder
	source, err := iofs.New(migrationFiles, migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration source %w", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{
		DatabaseName: dbName,
	})
	if err != nil {
		return nil, fmt.Errorf("creating sqlite3 db driver failed %s", err)
	}

	// Create a new instance of the migration using the defined source and target
	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance %w", err)
	}

	return m, nil
}
