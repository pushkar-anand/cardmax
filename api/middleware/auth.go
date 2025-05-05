package middleware

import (
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/auth"
	appcontext "github.com/pushkar-anand/cardmax/internal/context"
	"net/http"

	"github.com/pushkar-anand/build-with-go/http/response"
	"log/slog"
)

// AuthMiddleware creates a middleware that checks for a valid user session.
// It extracts the user ID from the session and adds it to the request context.
func AuthMiddleware(store *auth.SessionStore, log *slog.Logger, jw *response.JSONWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Get Session
			session, err := store.GetSession(r) // Use the same session name as in login
			if err != nil {
				log.WarnContext(ctx, "Failed to get session in auth middleware", logger.Error(err), slog.String("remote_addr", r.RemoteAddr))
				jw.WriteError(ctx, r, w, err)
				return
			}

			// 2. Retrieve User ID from Session Values
			userIDRaw := session.Values["user_id"]
			if userIDRaw == nil {
				log.DebugContext(ctx, "Auth middleware: user_id not found in session", slog.String("remote_addr", r.RemoteAddr))
				problem := response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Authentication required (session invalid).").Build()
				jw.WriteProblem(ctx, r, w, problem)
				return
			}

			// 3. Type Assert User ID
			userID, ok := userIDRaw.(int64)
			if !ok || userID <= 0 {
				log.WarnContext(ctx, "Auth middleware: Invalid user_id type or value in session", slog.Any("user_id_raw", userIDRaw), slog.String("remote_addr", r.RemoteAddr))
				problem := response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Authentication required (session corrupted).").Build()
				jw.WriteProblem(ctx, r, w, problem)
				return
			}

			ctx = appcontext.Add(ctx, appcontext.KeyUserID, userID)

			// 5. Call Next Handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
