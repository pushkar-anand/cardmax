package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/build-with-go/http/server"
	projectconfig "github.com/pushkar-anand/cardmax/config"
	"github.com/pushkar-anand/cardmax/internal/db"
	"github.com/pushkar-anand/cardmax/web"
	"log/slog"
)

func NewServer(
	cfg projectconfig.Server,
	logger *slog.Logger,
	tr *web.Renderer,
	jsonWriter *response.JSONWriter,
	reader *request.Reader,
	db *db.DB,
) *server.Server {
	h := mux.NewRouter()

	addRoutes(
		h,
		logger,
		tr,
		jsonWriter,
		reader,
	)

	s := server.New(
		h,
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithLogger(logger),
	)

	return s
}

func Serve(
	ctx context.Context,
	srv *server.Server,
) error {
	err := srv.Serve(ctx)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
