package users

import (
	"database/sql"
	"errors"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/auth"
	"log/slog"
	"net/http"
)

// CreateUserHandler handles user registration
func CreateUserHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
	repo Repository,
) http.HandlerFunc {
	type (
		Request struct {
			Username string `schema:"username" validate:"required,min=3"`
			Password string `schema:"password" validate:"required,min=8,max=1000"`
		}

		Response struct {
			User *User `json:"user"`
		}
	)

	typedReader := request.NewTypedReader[Request](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := typedReader.ReadAndValidateForm(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse request body", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		// Check if user already exists
		exists, err := repo.UserExists(ctx, body.Username)
		if err != nil {
			log.ErrorContext(ctx, "Failed to check for existing username", slog.String("username", body.Username), logger.Error(err))
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to process registration.").Build())
			return
		}

		if exists {
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusConflict).WithDetail("Username already taken.").Build())
			return
		}

		// Create user
		user, err := repo.CreateUser(ctx, body.Username, body.Password)
		if err != nil {
			log.ErrorContext(ctx, "failed to create user", logger.Error(err))
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to create user. Please try again.").Build())
			return
		}

		log.DebugContext(ctx, "user created", slog.String("username", body.Username), slog.Int64("user_id", user.ID))

		jw.Write(ctx, w, http.StatusCreated, Response{user})
	}
}

// LoginHandler handles user login
func LoginHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
	repo Repository,
	sessionStore *auth.SessionStore,
) http.HandlerFunc {
	type (
		Request struct {
			Username string `schema:"username" validate:"required,min=3"`
			Password string `schema:"password" validate:"required,min=8,max=1000"`
		}

		Response struct {
			User *User `json:"user"`
		}
	)

	typedReader := request.NewTypedReader[Request](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := typedReader.ReadAndValidateForm(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse request body", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		// Validate credentials
		user, err := repo.ValidateCredentials(ctx, body.Username, body.Password)
		if err != nil {
			var statusCode int
			var detail string

			if errors.Is(err, auth.ErrMismatchedHashAndPassword) {
				statusCode = http.StatusUnauthorized
				detail = "Invalid username or password."
				log.WarnContext(ctx, "Login attempt failed", slog.String("username", body.Username))
			} else if errors.Is(err, sql.ErrNoRows) {
				statusCode = http.StatusUnauthorized
				detail = "Invalid username or password."
				log.WarnContext(ctx, "Login attempt for non-existent user", slog.String("username", body.Username))
			} else {
				statusCode = http.StatusInternalServerError
				detail = "Login failed. Please try again."
				log.ErrorContext(ctx, "Login attempt error", slog.String("username", body.Username), logger.Error(err))
			}

			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(statusCode).WithDetail(detail).Build())
			return
		}

		// Create session
		err = sessionStore.New(w, r, map[string]any{
			"user_id":  user.ID,
			"username": user.Username,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to create session", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		log.InfoContext(ctx, "User logged in successfully", slog.String("username", user.Username), slog.Int64("userID", user.ID))

		jw.Write(ctx, w, http.StatusOK, Response{user})
	}
}
