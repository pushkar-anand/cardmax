package users

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for in-memory DB
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/internal/db"
	"github.com/pushkar-anand/cardmax/internal/db/models"
)

// testDeps holds dependencies needed for handler tests
type testDeps struct {
	handler *UserHandler
	dbConn  *sql.DB
	store   sessions.Store
	router  *mux.Router
}

// setupTest creates an in-memory DB, migrates it, and sets up handlers/router.
func setupTest(t *testing.T) *testDeps {
	t.Helper()

	// Use in-memory SQLite DB for testing
	// ":memory:" doesn't work well with multiple connections needed for migrations,
	// so use a temporary file that gets cleaned up.
	dbPath := "file::memory:?cache=shared" // In-memory DB shared across connections
	// Alternative: temp file
	// tempFile, err := os.CreateTemp("", "test_*.db")
	// if err != nil {
	// 	t.Fatalf("Failed to create temp db file: %v", err)
	// }
	// dbPath := tempFile.Name()
	// t.Cleanup(func() { os.Remove(dbPath) }) // Ensure DB file cleanup

	// Initialize DB connection
	// Need a separate logger for DB setup
	testLog := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})) // Or io.DiscardHandler
	dbConn, err := db.New(context.Background(), testLog, &db.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	err = dbConn.MigrateUp(testLog)
	if err != nil {
		dbConn.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup dependencies
	jw := response.NewJSONWriter(testLog)
	store := sessions.NewCookieStore([]byte("test-secret-key")) // Use a fixed key for tests

	handler := NewUserHandler(testLog, jw, dbConn, store)
	router := mux.NewRouter()

	// Register routes (mirroring routes.go)
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/users/register", handler.RegisterUserHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/users/login", handler.LoginUserHandler).Methods(http.MethodPost)

	t.Cleanup(func() {
		err := dbConn.Close()
		if err != nil {
			t.Errorf("Failed to close test DB connection: %v", err)
		}
	})

	return &testDeps{
		handler: handler,
		dbConn:  dbConn,
		store:   store,
		router:  router,
	}
}

// TestRegisterUserHandler tests the registration endpoint.
func TestRegisterUserHandler(t *testing.T) {
	deps := setupTest(t)

	testCases := []struct {
		name           string
		payload        map[string]string
		setupDB        func() // Optional setup specific to this test case (e.g., pre-create user)
		expectedStatus int
		expectBody     bool // Whether to check for specific fields in the response body
		expectedUser   string // Expected username in body if expectBody is true
	}{
		{
			name: "Successful registration",
			payload: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusCreated,
			expectBody:     true,
			expectedUser:   "testuser",
		},
		{
			name: "Duplicate username",
			payload: map[string]string{
				"username": "existinguser",
				"password": "password123",
			},
			setupDB: func() {
				// Pre-create the user
				_, err := deps.handler.DB.ExecContext(context.Background(),
					"INSERT INTO users (username, hashed_password) VALUES (?, ?)",
					"existinguser", "$2a$10$dummyhash") // Hash doesn't matter here
				if err != nil {
					t.Fatalf("Failed to setup duplicate user: %v", err)
				}
			},
			expectedStatus: http.StatusConflict,
			expectBody:     false,
		},
		{
			name:           "Empty username",
			payload:        map[string]string{"username": "", "password": "password123"},
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name:           "Empty password",
			payload:        map[string]string{"username": "newuser", "password": ""},
			expectedStatus: http.StatusBadRequest, // Assuming password length check catches this
			expectBody:     false,
		},
		{
			name:           "Short password",
			payload:        map[string]string{"username": "newuser2", "password": "short"},
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name:           "Malformed JSON",
			payload:        nil, // Special case handled below
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupDB != nil {
				tc.setupDB()
			}

			var reqBody []byte
			var err error
			if tc.name == "Malformed JSON" {
				reqBody = []byte(`{"username": "test", "password": "pass"`) // Missing closing brace
			} else {
				reqBody, err = json.Marshal(tc.payload)
				if err != nil {
					t.Fatalf("Failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/users/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			deps.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			if tc.expectBody && rr.Code == http.StatusCreated {
				var respBody RegisterResponse
				err := json.Unmarshal(rr.Body.Bytes(), &respBody)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v. Body: %s", err, rr.Body.String())
				}
				if respBody.Username != tc.expectedUser {
					t.Errorf("Expected username %q in response, got %q", tc.expectedUser, respBody.Username)
				}
				if respBody.ID == 0 {
					t.Errorf("Expected non-zero user ID in response")
				}
			}
		})
	}
}

// TestLoginUserHandler tests the login endpoint.
func TestLoginUserHandler(t *testing.T) {
	deps := setupTest(t)
	password := "password123"
	username := "loginuser"

	// Create a user to test login
	hashedPassword, err := deps.handler.HashPassword(password) // Using handler's HashPassword
	if err != nil {
		t.Fatalf("Setup failed: could not hash password: %v", err)
	}
	_, err = deps.dbConn.ExecContext(context.Background(),
		"INSERT INTO users (username, hashed_password) VALUES (?, ?)",
		username, hashedPassword)
	if err != nil {
		t.Fatalf("Setup failed: could not insert test user: %v", err)
	}

	testCases := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		expectCookie   bool
		expectBody     bool
	}{
		{
			name:           "Successful login",
			payload:        map[string]string{"username": username, "password": password},
			expectedStatus: http.StatusOK,
			expectCookie:   true,
			expectBody:     true,
		},
		{
			name:           "Incorrect password",
			payload:        map[string]string{"username": username, "password": "wrongpassword"},
			expectedStatus: http.StatusUnauthorized,
			expectCookie:   false,
			expectBody:     false,
		},
		{
			name:           "Non-existent username",
			payload:        map[string]string{"username": "nosuchuser", "password": password},
			expectedStatus: http.StatusUnauthorized,
			expectCookie:   false,
			expectBody:     false,
		},
		{
			name:           "Empty username",
			payload:        map[string]string{"username": "", "password": password},
			expectedStatus: http.StatusBadRequest,
			expectCookie:   false,
			expectBody:     false,
		},
		{
			name:           "Empty password",
			payload:        map[string]string{"username": username, "password": ""},
			expectedStatus: http.StatusBadRequest,
			expectCookie:   false,
			expectBody:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			deps.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			// Check for session cookie on successful login
			setCookieHeader := rr.Header().Get("Set-Cookie")
			hasCookie := strings.Contains(setCookieHeader, "user-session=")

			if tc.expectCookie && !hasCookie {
				t.Errorf("Expected 'user-session' cookie, but none found")
			}
			if !tc.expectCookie && hasCookie {
				t.Errorf("Did not expect 'user-session' cookie, but one was found: %s", setCookieHeader)
			}

			if tc.expectBody && rr.Code == http.StatusOK {
				var respBody LoginResponse
				err := json.Unmarshal(rr.Body.Bytes(), &respBody)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v. Body: %s", err, rr.Body.String())
				}
				if respBody.Username != username {
					t.Errorf("Expected username %q in response, got %q", username, respBody.Username)
				}
				if respBody.ID == 0 {
					t.Errorf("Expected non-zero user ID in response")
				}
				if respBody.Message != "Login successful" {
					t.Errorf("Expected message 'Login successful', got %q", respBody.Message)
				}
			}
		})
	}
}
