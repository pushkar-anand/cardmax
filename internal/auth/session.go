package auth

import (
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"time"
)

func init() {
	// Register types that will be stored in sessions
	// This allows complex types to be stored in the session
	gob.Register(map[string]interface{}{})
}

const sessionName = "cardmaxx"

var ErrNoSession = errors.New("no session")

// SessionStore wraps the gorilla sessions.Store interface
// to provide a more convenient API for working with sessions
type SessionStore struct {
	store *sessions.CookieStore
}

// SessionOptions holds configuration options for sessions
type SessionOptions struct {
	// Path for the cookie, defaults to "/"
	Path string
	// MaxAge for the cookie in seconds, defaults to 7 days
	MaxAge int
	// HttpOnly prevents JavaScript access to the cookie
	HttpOnly bool
	// Secure requires HTTPS connection
	Secure bool
	// SameSite controls cookie SameSite attribute (none, lax, strict)
	SameSite http.SameSite
	// Domain specifies the cookie domain, if empty the current domain is used
	Domain string
}

// NewSessionStore creates a new SessionStore with the provided cookie store
func NewSessionStore(secret string, options *SessionOptions) *SessionStore {
	// Create a new cookie store to handle session cookies.
	// We are not using the sessions.NewCookieStore() to create the store for use as it sets a default max age of 86400 * 30 seconds.
	// Instead, we are creating the store manually and setting the max age to whatever is passed in the options.
	//
	// GitHub issue:https://github.com/gorilla/securecookie/issues/44#issuecomment-1109603828
	store := &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs([]byte(secret)),
		Options: &sessions.Options{
			Path:     options.Path,
			Domain:   options.Domain,
			MaxAge:   options.MaxAge,
			Secure:   options.Secure,
			HttpOnly: options.HttpOnly,
			SameSite: options.SameSite,
		},
	}

	store.Options = &sessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	}

	return &SessionStore{
		store: store,
	}
}

func (s *SessionStore) New(w http.ResponseWriter, r *http.Request, data map[string]any) error {
	session, err := s.store.New(r, sessionName)
	if err != nil {
		return fmt.Errorf("SessionStore.New: failed to create session: %w", err)
	}

	for k, v := range data {
		session.Values[k] = v
	}

	return session.Save(r, w)
}

// GetSession retrieves the session from the request
func (s *SessionStore) GetSession(r *http.Request) (*sessions.Session, error) {
	return s.store.Get(r, sessionName)
}

// SetValue sets a value in the session
func (s *SessionStore) SetValue(r *http.Request, w http.ResponseWriter, key string, value interface{}) error {
	session, err := s.GetSession(r)
	if err != nil {
		return err
	}

	session.Values[key] = value
	return session.Save(r, w)
}

// GetValue gets a value from the session
func (s *SessionStore) GetValue(r *http.Request, key string) (interface{}, bool, error) {
	session, err := s.GetSession(r)
	if err != nil {
		return nil, false, err
	}

	val, exists := session.Values[key]

	return val, exists, nil
}

// GetInt64 gets an int64 value from the session
func (s *SessionStore) GetInt64(r *http.Request, key string) (int64, bool, error) {
	val, exists, err := s.GetValue(r, key)
	if err != nil || !exists {
		return 0, exists, err
	}

	intVal, ok := val.(int64)
	return intVal, ok, nil
}

// GetString gets a string value from the session
func (s *SessionStore) GetString(r *http.Request, key string) (string, bool, error) {
	val, exists, err := s.GetValue(r, key)
	if err != nil || !exists {
		return "", exists, err
	}

	strVal, ok := val.(string)
	return strVal, ok, nil
}

// ClearSession invalidates the session
func (s *SessionStore) ClearSession(r *http.Request, w http.ResponseWriter) error {
	session, err := s.GetSession(r)
	if err != nil {
		return err
	}

	// Clear all values
	for key := range session.Values {
		delete(session.Values, key)
	}

	// Set MaxAge to -1 to delete the cookie
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

// SetSessionExpiry sets the expiry time for the session
func (s *SessionStore) SetSessionExpiry(r *http.Request, w http.ResponseWriter, duration time.Duration) error {
	session, err := s.GetSession(r)
	if err != nil {
		return err
	}

	session.Options.MaxAge = int(duration.Seconds())
	return session.Save(r, w)
}

// Store returns the underlying gorilla sessions.Store
func (s *SessionStore) Store() sessions.Store {
	return s.store
}
