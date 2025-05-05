package main

import (
	"github.com/gorilla/mux"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"database/sql" // Need this for DB dependency
	"github.com/gorilla/mux"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/api/cards"
	"github.com/pushkar-anand/cardmax/api/middleware" // Import middleware
	"github.com/pushkar-anand/cardmax/api/recommend"
	"github.com/gorilla/sessions" // Import sessions
	"github.com/pushkar-anand/cardmax/api/users" // Import the users package
	"github.com/pushkar-anand/cardmax/web"
	"log/slog"
	"net/http"
)

// Assume db is accessible in this scope, e.g., passed in or global
var db *sql.DB // Placeholder assumption - real app should manage DB connection properly

func addRoutes(
	router *mux.Router,
	logger *slog.Logger,
	tr *web.Renderer,
	jsonWriter *response.JSONWriter,
	reader *request.Reader,
	dbConn *sql.DB, // Pass DB connection explicitly
	store sessions.Store, // Pass session store explicitly
) {
	// Update placeholder assumption
	db = dbConn // Assign passed DB connection

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

	// Public API routes (Login/Register)
	userHandler := users.NewUserHandler(logger, jsonWriter, db, store)
	apiRouter.HandleFunc("/users/register", userHandler.RegisterUserHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/users/login", userHandler.LoginUserHandler).Methods(http.MethodPost)

	// Authenticated API routes
	authenticatedAPIRouter := apiRouter.PathPrefix("").Subrouter() // Create a subrouter for authenticated routes
	authenticatedAPIRouter.Use(authMw)                             // Apply middleware to this subrouter

	// Instantiate handlers that need DB/Auth context (if not already done)
	cardHandler := cards.NewCardHandler(logger, jsonWriter, db) // Assuming CardHandler exists from previous steps
	// recommendHandler := recommend.NewRecommendHandler(...) // Assuming RecommendHandler exists

	// Apply authenticated routes to the subrouter
	// Existing Card routes (assuming they are user-specific now)
	authenticatedAPIRouter.HandleFunc(
		"/cards",
		cardHandler.ListCardsHandler, // Assuming ListCardsHandler exists and uses userID from context
	).Methods(http.MethodGet)
	authenticatedAPIRouter.HandleFunc(
		"/cards",
		cardHandler.CreateCardHandler, // Assuming CreateCardHandler exists and uses userID from context
	).Methods(http.MethodPost)
	authenticatedAPIRouter.HandleFunc(
		"/cards/{cardID}", // Assuming route uses cardID now
		cardHandler.GetCardHandler, // Assuming GetCardHandler exists and uses userID from context
	).Methods(http.MethodGet)
	authenticatedAPIRouter.HandleFunc(
		"/cards/{cardID}",
		cardHandler.UpdateCardHandler, // Assuming UpdateCardHandler exists and uses userID from context
	).Methods(http.MethodPut, http.MethodPatch) // Allow PUT/PATCH for updates
	authenticatedAPIRouter.HandleFunc(
		"/cards/{cardID}",
		cardHandler.DeleteCardHandler, // Assuming DeleteCardHandler exists and uses userID from context
	).Methods(http.MethodDelete)

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
