package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pushkar-anand/build-with-go/http/server"
	projectconfig "github.com/pushkar-anand/cardmax/config"
	"github.com/pushkar-anand/cardmax/web"
	"log/slog"
	"net/http"
)

func Serve(
	ctx context.Context,
	cfg projectconfig.Server,
	log *slog.Logger,
) error {
	handler, err := buildAndGetRoutes(log)
	if err != nil {
		return fmt.Errorf("buildAndGetRoutes failed: %w", err)
	}

	s := server.New(
		handler,
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithLogger(log),
	)

	err = s.Serve(ctx)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func buildAndGetRoutes(
	log *slog.Logger,
) (http.Handler, error) {
	templates, err := web.GetTemplates()
	if err != nil {
		return nil, fmt.Errorf("web.GetTemplates failed: %w", err)
	}

	h := mux.NewRouter()

	h.PathPrefix("/static/").Handler(web.StaticFilesHandler()).Methods(http.MethodGet)

	// Route handlers
	h.HandleFunc("/", templates.Handler(web.TemplateHome, nil)).Methods(http.MethodGet)
	h.HandleFunc("/cards", templates.Handler(web.TemplateCards, nil)).Methods(http.MethodGet)
	h.HandleFunc("/recommend", templates.Handler(web.TemplateRecommend, nil)).Methods(http.MethodGet)
	h.HandleFunc("/transactions", templates.Handler(web.TemplateTransactions, nil)).Methods(http.MethodGet)

	// API endpoints
	//h.HandleFunc("/api/predefined-cards", predefinedCardsHandler)

	return h, nil
}
