package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pushkar-anand/build-with-go/logger"
	"log/slog"
	"path"
	"sync"
)

//go:generate go tool sqlc generate

type (
	Config struct {
		Path string
	}

	DB struct {
		db *sql.DB
	}
)

const dbName = "cardmax.db"

var (
	once   sync.Once
	dbConn *DB
	dbErr  error
)

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *Config,
) (*DB, error) {
	var (
		db  *sql.DB
		err error
	)

	once.Do(func() {
		dbPath := path.Join(cfg.Path, dbName)

		db, err = connect(ctx, log, dbPath)
		if err != nil {
			log.ErrorContext(ctx, "failed to connect to database", logger.Error(err))
			dbErr = err
			return
		}

		dbConn = &DB{db: db}

		err = migrateDB(ctx, log, dbPath)
		if err != nil {
			log.ErrorContext(ctx, "failed to run migrations", logger.Error(err))
			dbErr = err
		}
	})

	return dbConn, dbErr
}

func connect(
	ctx context.Context,
	log *slog.Logger,
	dbName string,
) (*sql.DB, error) {
	open, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	log.InfoContext(ctx, "connected to SQLITE database", slog.String("path", dbName))

	return open, nil
}
