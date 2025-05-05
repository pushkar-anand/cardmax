package cards

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for in-memory DB
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/api/middleware" // Import middleware
	"github.com/pushkar-anand/cardmax/internal/auth"   // Import auth for password hashing
	"github.com/pushkar-anand/cardmax/internal/db"
	"github.com/pushkar-anand/cardmax/internal/db/models"
)

// --- Test Setup ---

// testCardDeps holds dependencies for card handler tests
type testCardDeps struct {
	cardHandler *CardHandler
	dbConn      *sql.DB
	store       sessions.Store
	router      *mux.Router
	log         *slog.Logger
	jw          *response.JSONWriter
}

// setupCardTest creates an in-memory DB, migrates it, sets up handlers, middleware, and router.
func setupCardTest(t *testing.T) *testCardDeps {
	t.Helper()

	dbPath := "file:card_test_" + t.Name() + ".db?cache=shared&mode=memory"
	// Clean up DB file after test
	t.Cleanup(func() {
		// Closing the connection should release the memory DB
	})

	testLog := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})) // Less verbose logging
	dbConn, err := db.New(context.Background(), testLog, &db.Config{Path: dbPath})
	require.NoError(t, err, "Failed to connect to test database")

	err = dbConn.MigrateUp(testLog)
	require.NoError(t, err, "Failed to run migrations")

	t.Cleanup(func() {
		assert.NoError(t, dbConn.Close(), "Failed to close test DB connection")
	})

	jw := response.NewJSONWriter(testLog)
	store := sessions.NewCookieStore([]byte("card-test-secret-key"))
	cardHandler := NewCardHandler(testLog, jw, dbConn)
	authMw := middleware.AuthMiddleware(store, testLog, jw)
	router := mux.NewRouter()

	// Setup authenticated routes (mirroring routes.go structure)
	apiRouter := router.PathPrefix("/api").Subrouter()
	authenticatedAPIRouter := apiRouter.PathPrefix("").Subrouter()
	authenticatedAPIRouter.Use(authMw)

	authenticatedAPIRouter.HandleFunc("/cards", cardHandler.ListCardsHandler).Methods(http.MethodGet)
	authenticatedAPIRouter.HandleFunc("/cards", cardHandler.CreateCardHandler).Methods(http.MethodPost)
	authenticatedAPIRouter.HandleFunc("/cards/{cardID}", cardHandler.GetCardHandler).Methods(http.MethodGet)
	authenticatedAPIRouter.HandleFunc("/cards/{cardID}", cardHandler.UpdateCardHandler).Methods(http.MethodPut, http.MethodPatch)
	authenticatedAPIRouter.HandleFunc("/cards/{cardID}", cardHandler.DeleteCardHandler).Methods(http.MethodDelete)

	return &testCardDeps{
		cardHandler: cardHandler,
		dbConn:      dbConn,
		store:       store,
		router:      router,
		log:         testLog,
		jw:          jw,
	}
}

// --- Helper Functions ---

// createUser creates a user directly in the DB and returns the user ID.
func (d *testCardDeps) createUser(t *testing.T, username, password string) int64 {
	t.Helper()
	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err, "Failed to hash password for test user")

	result, err := d.dbConn.ExecContext(context.Background(),
		"INSERT INTO users (username, hashed_password) VALUES (?, ?)",
		username, hashedPassword)
	require.NoError(t, err, "Failed to insert test user")

	userID, err := result.LastInsertId()
	require.NoError(t, err, "Failed to get last insert ID for user")
	require.NotZero(t, userID, "User ID should not be zero")
	return userID
}

// createCard creates a card directly in the DB for a given user and returns the card ID.
func (d *testCardDeps) createCard(t *testing.T, userID int64, name, issuer string) int64 {
	t.Helper()
	params := models.CreateCardParams{
		UserID:      userID,
		Name:        name,
		Issuer:      issuer,
		Last4Digits: "1234",
		ExpiryDate:  "12/25",
		CardType:    "Visa",
	}
	queries := models.New(d.dbConn)
	card, err := queries.CreateCard(context.Background(), params)
	require.NoError(t, err, "Failed to insert test card")
	require.NotZero(t, card.ID, "Card ID should not be zero")
	return card.ID
}

// createAuthenticatedRequest creates an HTTP request with a valid session cookie for the given user ID.
func (d *testCardDeps) createAuthenticatedRequest(t *testing.T, method, path string, body []byte, userID int64) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create and save session to get cookie
	session, _ := d.store.Get(req, "user-session")
	session.Values["user_id"] = userID
	rr := httptest.NewRecorder() // Dummy recorder to save session
	err := session.Save(req, rr)
	require.NoError(t, err, "Failed to save session for request")

	// Set cookie on the actual request
	cookieHeader := rr.Header().Get("Set-Cookie")
	require.NotEmpty(t, cookieHeader, "Set-Cookie header missing")
	req.Header.Set("Cookie", cookieHeader)

	return req
}

// --- Test Cases ---

// TestCardHandlers_Unauthenticated tests accessing card endpoints without authentication.
func TestCardHandlers_Unauthenticated(t *testing.T) {
	deps := setupCardTest(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/cards"},
		{http.MethodPost, "/api/cards"},
		{http.MethodGet, "/api/cards/1"},
		{http.MethodPut, "/api/cards/1"},
		{http.MethodDelete, "/api/cards/1"},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s_%s", ep.method, strings.ReplaceAll(ep.path, "/", "_")), func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil) // No session cookie
			rr := httptest.NewRecorder()
			deps.router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected Unauthorized status")
		})
	}
}

// TestCardHandlers_CRUD tests the full lifecycle (Create, Read, Update, Delete) for cards.
func TestCardHandlers_CRUD(t *testing.T) {
	deps := setupCardTest(t)

	// Create a test user
	userID := deps.createUser(t, "carduser", "password")

	// --- Create Card ---
	var createdCardID int64
	t.Run("Create Card", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "My Test Card",
			"issuer":      "Test Bank",
			"last4_digits": "5678",
			"expiry_date": "01/26",
			"card_type":   "Mastercard",
		}
		body, _ := json.Marshal(payload)
		req := deps.createAuthenticatedRequest(t, http.MethodPost, "/api/cards", body, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "Expected OK status on create") // sqlc returns 200 OK by default

		var respCard models.Card
		err := json.Unmarshal(rr.Body.Bytes(), &respCard)
		require.NoError(t, err, "Failed to unmarshal create response")
		assert.Equal(t, payload["name"], respCard.Name)
		assert.Equal(t, payload["issuer"], respCard.Issuer)
		assert.Equal(t, userID, respCard.UserID)
		require.NotZero(t, respCard.ID)
		createdCardID = respCard.ID // Save for later steps
	})

	require.NotZero(t, createdCardID, "Created card ID should be set for subsequent tests")

	// --- List Cards ---
	t.Run("List Cards", func(t *testing.T) {
		// Create another card for the same user
		deps.createCard(t, userID, "Second Card", "Another Bank")

		// Create a card for a different user (should not be listed)
		otherUserID := deps.createUser(t, "otheruser", "password")
		deps.createCard(t, otherUserID, "Other User Card", "Bank Z")

		req := deps.createAuthenticatedRequest(t, http.MethodGet, "/api/cards", nil, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "Expected OK status on list")

		var resp struct {
			Cards []models.Card `json:"cards"`
		}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err, "Failed to unmarshal list response")

		assert.Len(t, resp.Cards, 2, "Expected 2 cards for the logged-in user")
		foundFirst := false
		foundSecond := false
		for _, card := range resp.Cards {
			assert.Equal(t, userID, card.UserID, "Listed card belongs to wrong user")
			if card.ID == createdCardID {
				foundFirst = true
			}
			if card.Name == "Second Card" {
				foundSecond = true
			}
		}
		assert.True(t, foundFirst, "Original created card not found in list")
		assert.True(t, foundSecond, "Second created card not found in list")
	})

	// --- Get Card ---
	t.Run("Get Card - Success", func(t *testing.T) {
		path := fmt.Sprintf("/api/cards/%d", createdCardID)
		req := deps.createAuthenticatedRequest(t, http.MethodGet, path, nil, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "Expected OK status on get")

		var respCard models.Card
		err := json.Unmarshal(rr.Body.Bytes(), &respCard)
		require.NoError(t, err, "Failed to unmarshal get response")
		assert.Equal(t, createdCardID, respCard.ID)
		assert.Equal(t, "My Test Card", respCard.Name)
		assert.Equal(t, userID, respCard.UserID)
	})

	t.Run("Get Card - Not Found", func(t *testing.T) {
		path := "/api/cards/99999" // Non-existent ID
		req := deps.createAuthenticatedRequest(t, http.MethodGet, path, nil, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "Expected Not Found status")
	})

	// --- Update Card ---
	t.Run("Update Card - Success", func(t *testing.T) {
		path := fmt.Sprintf("/api/cards/%d", createdCardID)
		payload := map[string]interface{}{
			"name":        "Updated Test Card", // Changed name
			"issuer":      "Updated Bank",      // Changed issuer
			"last4_digits": "1111",
			"expiry_date": "11/27",
			"card_type":   "Visa",
		}
		body, _ := json.Marshal(payload)
		req := deps.createAuthenticatedRequest(t, http.MethodPut, path, body, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code, "Expected OK status on update")

		var respCard models.Card
		err := json.Unmarshal(rr.Body.Bytes(), &respCard)
		require.NoError(t, err, "Failed to unmarshal update response")
		assert.Equal(t, createdCardID, respCard.ID)
		assert.Equal(t, "Updated Test Card", respCard.Name) // Verify updated name
		assert.Equal(t, "Updated Bank", respCard.Issuer)    // Verify updated issuer
		assert.Equal(t, userID, respCard.UserID)

		// Optional: Verify in DB
		dbCard, err := models.New(deps.dbConn).GetCardByIDAndUser(context.Background(), models.GetCardByIDAndUserParams{ID: createdCardID, UserID: userID})
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Card", dbCard.Name)
	})

	t.Run("Update Card - Not Found", func(t *testing.T) {
		path := "/api/cards/99999" // Non-existent ID
		payload := map[string]interface{}{"name": "Doesn't Matter"}
		body, _ := json.Marshal(payload)
		req := deps.createAuthenticatedRequest(t, http.MethodPut, path, body, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "Expected Not Found status")
	})

	// --- Delete Card ---
	t.Run("Delete Card - Success", func(t *testing.T) {
		path := fmt.Sprintf("/api/cards/%d", createdCardID)
		req := deps.createAuthenticatedRequest(t, http.MethodDelete, path, nil, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code, "Expected No Content status on delete")

		// Optional: Verify deletion in DB
		_, err := models.New(deps.dbConn).GetCardByIDAndUser(context.Background(), models.GetCardByIDAndUserParams{ID: createdCardID, UserID: userID})
		assert.ErrorIs(t, err, sql.ErrNoRows, "Card should not exist in DB after delete")
	})

	t.Run("Delete Card - Not Found", func(t *testing.T) {
		path := fmt.Sprintf("/api/cards/%d", createdCardID) // Already deleted
		req := deps.createAuthenticatedRequest(t, http.MethodDelete, path, nil, userID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		// Our DeleteCard query doesn't return error if not found, just affects 0 rows.
		// The handler returns 204 regardless in the current implementation.
		// A more robust implementation might check RowsAffected and return 404.
		assert.Equal(t, http.StatusNoContent, rr.Code, "Expected No Content status even if card was already deleted")

		// Test deleting a completely non-existent ID
		pathNotFound := "/api/cards/99999"
		reqNotFound := deps.createAuthenticatedRequest(t, http.MethodDelete, pathNotFound, nil, userID)
		rrNotFound := httptest.NewRecorder()
		deps.router.ServeHTTP(rrNotFound, reqNotFound)
		assert.Equal(t, http.StatusNoContent, rrNotFound.Code, "Expected No Content status for non-existent ID")

	})
}

// TestCardHandlers_Authorization tests that users cannot access each other's cards.
func TestCardHandlers_Authorization(t *testing.T) {
	deps := setupCardTest(t)

	// Create User A and their card
	userAID := deps.createUser(t, "userA", "passwordA")
	cardAID := deps.createCard(t, userAID, "Card A", "Bank A")

	// Create User B
	userBID := deps.createUser(t, "userB", "passwordB")

	endpointsToTest := []struct {
		name   string
		method string
		path   string // Path using User A's card ID
		body   []byte // Optional body for PUT/POST
	}{
		{"Get Card A", http.MethodGet, fmt.Sprintf("/api/cards/%d", cardAID), nil},
		{"Update Card A", http.MethodPut, fmt.Sprintf("/api/cards/%d", cardAID), []byte(`{"name":"Attempted Update"}`)},
		{"Delete Card A", http.MethodDelete, fmt.Sprintf("/api/cards/%d", cardAID), nil},
	}

	for _, ep := range endpointsToTest {
		t.Run(fmt.Sprintf("User B accessing %s", ep.name), func(t *testing.T) {
			// User B makes a request for User A's card
			req := deps.createAuthenticatedRequest(t, ep.method, ep.path, ep.body, userBID) // Authenticated as User B
			rr := httptest.NewRecorder()
			deps.router.ServeHTTP(rr, req)

			// Get/Update should return 404 because the query includes `WHERE id = ? AND user_id = ?`
			// Delete currently returns 204 even if no rows are affected.
			expectedStatus := http.StatusNotFound
			if ep.method == http.MethodDelete {
				expectedStatus = http.StatusNoContent // Current behavior
			}

			assert.Equal(t, expectedStatus, rr.Code, "Expected status %d when User B accesses User A's card", expectedStatus)

			// Verify User A's card still exists and wasn't modified/deleted by User B
			_, err := models.New(deps.dbConn).GetCardByIDAndUser(context.Background(), models.GetCardByIDAndUserParams{ID: cardAID, UserID: userAID})
			assert.NoError(t, err, "User A's card should still exist after User B's failed attempt")
		})
	}

	// Test List Cards - User B should get an empty list
	t.Run("User B List Cards", func(t *testing.T) {
		req := deps.createAuthenticatedRequest(t, http.MethodGet, "/api/cards", nil, userBID)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Cards []models.Card `json:"cards"`
		}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Empty(t, resp.Cards, "User B should have no cards listed")
	})
}
