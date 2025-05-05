package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pushkar-anand/build-with-go/http/response" // Assuming this provides WriteProblem or similar
	"log/slog" // For logging within middleware if needed
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// UserIDKey is the key used to store the user ID in the request context.
const UserIDKey contextKey = "user_id"

// AuthMiddleware creates a middleware that checks for a valid user session.
// It extracts the user ID from the session and adds it to the request context.
func AuthMiddleware(store sessions.Store, log *slog.Logger, jw *response.JSONWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Get Session
			session, err := store.Get(r, "user-session") // Use the same session name as in login
			if err != nil {
				// Log the error but treat it as unauthorized as we can't verify the session.
				log.WarnContext(ctx, "Failed to get session in auth middleware", slog.Any("error", err), slog.String("remote_addr", r.RemoteAddr))
				// Use JSONWriter if available for consistent error responses
				if jw != nil {
					problem := response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Authentication required.").Build()
					jw.WriteProblem(ctx, r, w, problem)
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}
				return
			}

			// Check if session is new or empty (might indicate no login or expired)
			if session.IsNew {
					log.DebugContext(ctx, "Auth middleware: Session is new", slog.String("remote_addr", r.RemoteAddr))
					if jw != nil {
							problem := response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Authentication required.").Build()
							jw.WriteProblem(ctx, r, w, problem)
					} else {
							http.Error(w, "Unauthorized", http.StatusUnauthorized)
					}
					return
			}


			// 2. Retrieve User ID from Session Values
			userIDRaw := session.Values["user_id"]
			if userIDRaw == nil {
				log.DebugContext(ctx, "Auth middleware: user_id not found in session", slog.String("remote_addr", r.RemoteAddr))
				if jw != nil {
					problem := response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Authentication required (session invalid).").Build()
					jw.WriteProblem(ctx, r, w, problem)
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}
				return
			}

			// 3. Type Assert User ID
			userID, ok := userIDRaw.(int64)
			if !ok || userID <= 0 {
				log.WarnContext(ctx, "Auth middleware: Invalid user_id type or value in session", slog.Any("user_id_raw", userIDRaw), slog.String("remote_addr", r.RemoteAddr))
				// Consider clearing the invalid session value here?
				// delete(session.Values, "user_id")
				// session.Save(r, w)
				if jw != nil {
					problem := response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Authentication required (session corrupted).").Build()
					jw.WriteProblem(ctx, r, w, problem)
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}
				return
			}

			// 4. Add User ID to Context
			// Use our custom context key type
			ctx = context.WithValue(r.Context(), UserIDKey, userID)
			log.DebugContext(ctx, "Auth middleware: User authenticated", slog.Int64("userID", userID), slog.String("uri", r.RequestURI))


			// 5. Call Next Handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the user ID stored in the context by AuthMiddleware.
// It returns the user ID and true if found, otherwise 0 and false.
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok && userID > 0
}
