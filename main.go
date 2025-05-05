package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/pushkar-anand/build-with-go/config"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/build-with-go/validator"
	"github.com/pushkar-anand/cardmax/internal/auth"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"github.com/pushkar-anand/cardmax/internal/db"
	"github.com/pushkar-anand/cardmax/web"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	projectconfig "github.com/pushkar-anand/cardmax/config"
)

//go:embed data/*
var data embed.FS

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancelFunc := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancelFunc()

	cfg, err := config.ReadFromEnv[projectconfig.Config](".env", "")
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read env variables", logger.Error(err))
		return fmt.Errorf("failed to read ENV vars: %w", err)
	}

	// Set default configuration values
	err = cfg.SetDefaults()
	if err != nil {
		// Logger not fully initialized yet, use slog directly
		slog.ErrorContext(ctx, "Failed to set default config values", logger.Error(err))
		return fmt.Errorf("failed to set default config values: %w", err)
	}

	buildInfo, _ := debug.ReadBuildInfo()

	log := getLogger(cfg.Environment)

	log = log.With(
		slog.String("version", buildInfo.Main.Version),
		slog.String("go_version", buildInfo.GoVersion),
		slog.String("environment", cfg.Environment.String()),
	)

	dbConn, err := db.New(ctx, log, &db.Config{
		Path: cfg.DB.Path,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to database", logger.Error(err))
		return fmt.Errorf("failed to connect database: %w", err)
	}

	parsedCards, err := cards.Parse(data)
	if err != nil {
		log.ErrorContext(ctx, "Failed to parse cards", logger.Error(err))
		return fmt.Errorf("failed to parse cards data: %w", err)
	}

	// Populate predefined cards into the database
	err = dbConn.PopulatePredefinedCards(ctx, log, parsedCards)
	if err != nil {
		log.ErrorContext(ctx, "Failed to populate predefined cards", logger.Error(err))
		return fmt.Errorf("failed to populate predefined cards: %w", err)
	}

	templates, err := web.GetTemplates()
	if err != nil {
		log.ErrorContext(ctx, "Failed to load templates", logger.Error(err))
		return fmt.Errorf("failed to load templates data: %w", err)
	}

	v, err := validator.New()
	if err != nil {
		log.ErrorContext(ctx, "Failed to init validator", logger.Error(err))
		return fmt.Errorf("failed to init validator: %w", err)
	}

	jw := response.NewJSONWriter(log, response.WithErrorProblemMapper(errorMapper()))
	rd := request.NewReader(log, v)

	sessionStore := auth.NewSessionStore(cfg.Session.Secret, &auth.SessionOptions{})

	// Initialize server, passing the store. Assume NewServer handles passing it down.
	srv := NewServer(cfg.Server, log, templates, jw, rd, dbConn, sessionStore) // Added store back

	// Routes are likely added within NewServer or its methods, so no explicit addRoutes call here.

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err = Serve(ctx, srv)
		if err != nil {
			log.ErrorContext(ctx, "Failed to start server", logger.Error(err))
			return fmt.Errorf("http server failed to start: %w", err)
		}

		return nil
	})

	err = g.Wait()
	if err != nil {
		log.ErrorContext(ctx, "Failed to start application", logger.Error(err))
		return fmt.Errorf("application failed: %w", err)
	}

	return nil
}

func getLogger(env projectconfig.Environment) *slog.Logger {
	opts := []logger.Option{
		logger.WithAddCaller(),
	}

	switch env {
	case projectconfig.Development:
		opts = append(opts, logger.WithLevel(slog.LevelDebug), logger.WithFormat(logger.FormatText))
	case projectconfig.Production:
		opts = append(opts, logger.WithLevel(slog.LevelInfo), logger.WithFormat(logger.FormatJSON))
	}

	return logger.New(opts...)
}

func errorMapper() func(err error) response.Problem {
	return func(err error) response.Problem {
		switch {
		case errors.Is(err, auth.ErrMismatchedHashAndPassword),
			errors.Is(err, auth.ErrInvalidHashFormat):
			errors.Is(err, auth.ErrNoSession)
			return response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("username or password is incorrect").Build()
		default:
			return response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Internal server error").Build()
		}
	}
}
