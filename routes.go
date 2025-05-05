package main

import (
	"github.com/gorilla/mux"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/api/middleware" // Import middleware
	"github.com/pushkar-anand/cardmax/api/recommend"
	"github.com/pushkar-anand/cardmax/api/users" // Import the users package
	"github.com/pushkar-anand/cardmax/internal/auth"
	"github.com/pushkar-anand/cardmax/internal/db"
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
	dbConn *db.DB,
	store *auth.SessionStore,
) {
	// Instantiate Middleware
	authMw := middleware.AuthMiddleware(store, logger, jsonWriter)

	// --- Public Routes ---
	// Static files
	router.PathPrefix("/static/").Handler(web.StaticFilesHandler()).Methods(http.MethodGet)

	// Web Templates (Assuming public for now, might need auth later)
	router.HandleFunc("/", tr.HTMLHandler(web.TemplateHome)).Methods(http.MethodGet)
	router.HandleFunc("/cards", tr.HTMLHandler(web.TemplateCards)).Methods(http.MethodGet)
	router.HandleFunc("/recommend", tr.HTMLHandler(web.TemplateRecommend)).Methods(http.MethodGet)
	router.HandleFunc("/transactions", tr.HTMLHandler(web.TemplateTransactions)).Methods(http.MethodGet)

	// --- API Routes ---
	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc(
		"/users", users.CreateUserHandler(logger, jsonWriter, reader, dbConn.Queries),
	).Methods(http.MethodPost)
	apiRouter.HandleFunc(
		"/users/login", users.LoginHandler(logger, jsonWriter, reader, dbConn.Queries, store),
	).Methods(http.MethodPost)

	// Authenticated API routes
	authenticatedAPIRouter := apiRouter.PathPrefix("").Subrouter() // Create a subrouter for authenticated routes
	authenticatedAPIRouter.Use(authMw)                             // Apply middleware to this subrouter

	// Recommendation routes
	authenticatedAPIRouter.HandleFunc(
		"/recommend",
		recommend.GetRecommendationHandler(logger, jsonWriter, reader), // Needs update to use userID
	).Methods(http.MethodPost)
	authenticatedAPIRouter.HandleFunc(
		"/recommend-html",
		recommend.GetRecommendationHTMLHandler(logger, reader, tr), // Needs update to use userID
	).Methods(http.MethodPost)

	// Add other authenticated routes (e.g., transactions) here
}
