package auth

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSessionStore(t *testing.T) {
	// Test with default options
	store := NewSessionStore("test-secret", &SessionOptions{
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: false,
		Secure:   false,
		SameSite: 0,
		Domain:   "",
	})

	// Test with custom options
	customOpts := &SessionOptions{
		Path:     "/api",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   "example.com",
	}

	store = NewSessionStore("test-secret", customOpts)
	assert.Equal(t, "/api", store.store.Options.Path)
	assert.Equal(t, 3600, store.store.Options.MaxAge)
	assert.Equal(t, true, store.store.Options.HttpOnly)
	assert.Equal(t, true, store.store.Options.Secure)
	assert.Equal(t, http.SameSiteStrictMode, store.store.Options.SameSite)
	assert.Equal(t, "example.com", store.store.Options.Domain)
}

func TestSessionStoreBasicOperations(t *testing.T) {
	store := NewSessionStore("test-secret", &SessionOptions{
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: 0,
		Domain:   "example.com",
	})
	// Create a test request and response writer
	r := httptest.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()

	err := store.New(rr, r, map[string]any{
		"foo":      "bar",
		"test-key": "test-value",
	})
	require.NoError(t, err)

	// Check response cookies - should have set the cookie
	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1)

	// Create a new request with the cookie
	r2 := httptest.NewRequest("GET", "http://example.com", nil)
	r2.AddCookie(cookies[0])

	// Test getting the value back
	val, exists, err := store.GetValue(r2, "test-key")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "test-value", val)

	// Test string helper
	strVal, exists, err := store.GetString(r2, "test-key")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "test-value", strVal)

	// Test getting a non-existent value
	_, exists, _ = store.GetValue(r2, "non-existent")
	assert.False(t, exists)
}

func TestClearSession(t *testing.T) {
	store := NewSessionStore("test-secret", &SessionOptions{
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: 0,
		Domain:   "example.com",
	})

	// Create a test request and response writer
	r := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	// Set a value
	err := store.SetValue(r, w, "test-key", "test-value")
	assert.NoError(t, err)

	// Get cookies and create a new request
	cookies := w.Result().Cookies()
	r2 := httptest.NewRequest("GET", "http://example.com", nil)
	r2.AddCookie(cookies[0])

	// Verify the value exists
	_, exists, _ := store.GetValue(r2, "test-key")
	assert.True(t, exists)

	// Clear the session
	w2 := httptest.NewRecorder()
	err = store.ClearSession(r2, w2)
	assert.NoError(t, err)

	// Check the cookie was marked for deletion (MaxAge < 0)
	clearedCookies := w2.Result().Cookies()
	assert.Len(t, clearedCookies, 1)
	assert.Negative(t, clearedCookies[0].MaxAge)
}
