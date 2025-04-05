package main

import (
	"github.com/gorilla/mux"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/api/cards"
	"github.com/pushkar-anand/cardmax/web"
	"log/slog"
	"net/http"
)

func addRoutes(
	router *mux.Router,
	logger *slog.Logger,
	tr *web.Renderer,
	jsonWriter *response.JSONWriter,
	reader *request.Reader,
) {
	router.PathPrefix("/static/").Handler(web.StaticFilesHandler()).Methods(http.MethodGet)

	router.HandleFunc("/", tr.HTMLHandler(web.TemplateHome)).Methods(http.MethodGet)
	router.HandleFunc("/cards", tr.HTMLHandler(web.TemplateCards)).Methods(http.MethodGet)
	router.HandleFunc("/recommend", tr.HTMLHandler(web.TemplateRecommend)).Methods(http.MethodGet)
	router.HandleFunc("/transactions", tr.HTMLHandler(web.TemplateTransactions)).Methods(http.MethodGet)

	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc(
		"/cards",
		cards.GetAllHandler(logger, jsonWriter),
	).Methods(http.MethodGet)
	apiRouter.HandleFunc(
		"/cards/{key}",
		cards.GetByKeyHandler(logger, jsonWriter),
	).Methods(http.MethodGet)

}
