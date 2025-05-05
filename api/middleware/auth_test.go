package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/stretchr/testify/assert" // Using testify for assertions
)

// testMiddlewareDeps holds dependencies for middleware tests
type testMiddlewareDeps struct {
	store  sessions.Store
	router *mux.Router
	// Mock dependencies for AuthMiddleware
	log *slog.Logger
	jw  *response.JSONWriter
}

// setupMiddlewareTest sets up a router with public and protected routes.
func setupMiddlewareTest(t *testing.T) *testMiddlewareDeps {
	t.Helper()

	store := sessions.NewCookieStore([]byte("test-secret-key-middleware"))
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})) // Use Stderr or Discard
	jw := response.NewJSONWriter(log)
	router := mux.NewRouter()

	// Middleware instance
	authMw := AuthMiddleware(store, log, jw)

	// --- Test Handlers ---
	// Public handler
	publicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Public OK")
	})

	// Protected handler - checks for user ID in context
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			// This case should ideally be prevented by the middleware, but good to check
			log.ErrorContext(r.Context(), "Protected handler reached but user ID not found in context")
			http.Error(w, "Internal Server Error: User ID missing after auth", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Protected OK for user %d", userID)
	})

	// --- Routes ---
	router.Handle("/public", publicHandler).Methods(http.MethodGet)

	// Apply middleware to the protected route
	router.Handle("/protected", authMw(protectedHandler)).Methods(http.MethodGet)

	return &testMiddlewareDeps{
		store:  store,
		router: router,
		log:    log,
		jw:     jw,
	}
}

// TestAuthMiddleware tests the behavior of the authentication middleware.
func TestAuthMiddleware(t *testing.T) {
	deps := setupMiddlewareTest(t)

	// --- Test Case 1: Accessing Public Route ---
	t.Run("Access Public Route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public", nil)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Public route should return OK")
		assert.Contains(t, rr.Body.String(), "Public OK", "Public route response body mismatch")
	})

	// --- Test Case 2: Accessing Protected Route without Session ---
	t.Run("Access Protected Route without Session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rr := httptest.NewRecorder()
		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Protected route without session should return Unauthorized")
		// Check for JSON problem details if jw is used
		assert.Contains(t, rr.Body.String(), `"status":401`, "Protected route without session should return JSON problem details")
		assert.Contains(t, rr.Body.String(), "Authentication required", "Protected route without session response body mismatch")

	})

	// --- Test Case 3: Accessing Protected Route with Valid Session ---
	t.Run("Access Protected Route with Valid Session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rr := httptest.NewRecorder() // Need a new recorder for each subtest

		// Create a valid session
		session, _ := deps.store.Get(req, "user-session")
		testUserID := int64(123)
		session.Values["user_id"] = testUserID
		// Save session to get the cookie header value to set on the request
		// We need a ResponseWriter to save the session *to*, so we use a dummy recorder here first.
		dummyRecorder := httptest.NewRecorder()
		err := session.Save(req, dummyRecorder)
		assert.NoError(t, err, "Failed to save session for cookie setup")

		// Get the cookie from the dummy recorder and set it on the actual request
		cookieHeader := dummyRecorder.Header().Get("Set-Cookie")
		assert.NotEmpty(t, cookieHeader, "Set-Cookie header should not be empty after saving session")
		req.Header.Set("Cookie", cookieHeader)

		// Serve the request with the cookie
		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Protected route with valid session should return OK")
		expectedBody := fmt.Sprintf("Protected OK for user %d", testUserID)
		assert.Contains(t, rr.Body.String(), expectedBody, "Protected route response body mismatch")
	})

	// --- Test Case 4: Accessing Protected Route with Session missing user_id ---
	t.Run("Access Protected Route with Session missing user_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rr := httptest.NewRecorder()

		// Create a session but don't add user_id
		session, _ := deps.store.Get(req, "user-session")
		session.Values["other_data"] = "some value" // Add some other data to make sure it's not totally empty
		dummyRecorder := httptest.NewRecorder()
		err := session.Save(req, dummyRecorder)
		assert.NoError(t, err, "Failed to save session")
		req.Header.Set("Cookie", dummyRecorder.Header().Get("Set-Cookie"))

		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Protected route with session missing user_id should return Unauthorized")
		assert.Contains(t, rr.Body.String(), "session invalid", "Response body should indicate invalid session")

	})

	// --- Test Case 5: Accessing Protected Route with Invalid user_id type ---
	t.Run("Access Protected Route with Invalid user_id type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rr := httptest.NewRecorder()

		session, _ := deps.store.Get(req, "user-session")
		session.Values["user_id"] = "not-an-int64" // Invalid type
		dummyRecorder := httptest.NewRecorder()
		err := session.Save(req, dummyRecorder)
		assert.NoError(t, err, "Failed to save session")
		req.Header.Set("Cookie", dummyRecorder.Header().Get("Set-Cookie"))

		deps.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code, "Protected route with invalid user_id type should return Unauthorized")
		assert.Contains(t, rr.Body.String(), "session corrupted", "Response body should indicate corrupted session")
	})

	// --- Test Case 6: GetUserIDFromContext Helper ---
	t.Run("GetUserIDFromContext Helper", func(t *testing.T) {
		// Case A: Context without user ID
		ctx := context.Background()
		userID, ok := GetUserIDFromContext(ctx)
		assert.False(t, ok, "GetUserIDFromContext should return false when key is missing")
		assert.Zero(t, userID, "GetUserIDFromContext should return 0 when key is missing")

		// Case B: Context with valid user ID
		testUserID := int64(456)
		ctxWithID := context.WithValue(ctx, UserIDKey, testUserID)
		userID, ok = GetUserIDFromContext(ctxWithID)
		assert.True(t, ok, "GetUserIDFromContext should return true when key is present")
		assert.Equal(t, testUserID, userID, "GetUserIDFromContext returned wrong user ID")

		// Case C: Context with invalid type for user ID
		ctxWithWrongType := context.WithValue(ctx, UserIDKey, "wrong-type")
		userID, ok = GetUserIDFromContext(ctxWithWrongType)
		assert.False(t, ok, "GetUserIDFromContext should return false for wrong type")
		assert.Zero(t, userID, "GetUserIDFromContext should return 0 for wrong type")

		// Case D: Context with zero user ID (should be treated as invalid)
		ctxWithZeroID := context.WithValue(ctx, UserIDKey, int64(0))
		userID, ok = GetUserIDFromContext(ctxWithZeroID)
		assert.False(t, ok, "GetUserIDFromContext should return false for zero user ID")
		assert.Zero(t, userID, "GetUserIDFromContext should return 0 for zero user ID")
	})
}
