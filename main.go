package main

import (
	"context"
	"embed"
	"github.com/pushkar-anand/build-with-go/config"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	projectconfig "github.com/pushkar-anand/cardmax/config"
)

//go:embed data/
var data embed.FS

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancelFunc()

	cfg, err := config.ReadFromEnv[projectconfig.Config](".env", "")
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read env variables", logger.Error(err))
		panic(err)
	}

	buildInfo, _ := debug.ReadBuildInfo()

	log := getLogger(cfg.Environment)

	log = log.With(
		slog.String("version", buildInfo.Main.Version),
		slog.String("go_version", buildInfo.GoVersion),
		slog.String("environment", cfg.Environment.String()),
	)

	err = cards.Parse(data)
	if err != nil {
		log.ErrorContext(ctx, "Failed to parse cards", logger.Error(err))
		panic(err)
	}

	err = Serve(ctx, cfg.Server, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to start server", logger.Error(err))
		panic(err)
	}
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
