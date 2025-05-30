package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/db/models"
	"log/slog"
	"path"
	"strings"
	"sync"
)

//go:generate go tool sqlc generate

type (
	Config struct {
		Path string
	}

	DB struct {
		Conn    *sql.DB
		Queries *models.Queries
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
		dbPath := cfg.Path

		if !strings.HasSuffix(dbPath, ".db") {
			dbPath = path.Join(dbPath, dbName)
		}

		db, err = connect(ctx, log, dbPath)
		if err != nil {
			log.ErrorContext(ctx, "failed to connect to database", logger.Error(err))
			dbErr = err
			return
		}

		// Create the queries
		queries := models.New(db)

		dbConn = &DB{
			Conn:    db,
			Queries: queries,
		}

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
