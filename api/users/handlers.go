package users

import (
	"database/sql"
	"errors"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/auth"      // Import the auth package
	"github.com/pushkar-anand/cardmax/internal/db/models" // Import the generated models
	"log/slog"
	"net/http"
)

func CreateUserHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
	queries *models.Queries,
) http.HandlerFunc {
	type Request struct {
		Username string `schema:"username" validate:"required,min=3"`
		Password string `schema:"password" validate:"required,min=8,max=1000"`
	}

	typedReader := request.NewTypedReader[Request](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := typedReader.ReadAndValidateForm(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse request body", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		_, err = queries.GetUserByUsername(ctx, body.Username)
		if err == nil {
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusConflict).WithDetail("Username already taken.").Build())
			return
		} else if !errors.Is(err, sql.ErrNoRows) {
			log.ErrorContext(ctx, "Failed to check for existing username", slog.String("username", body.Username), logger.Error(err))
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to process registration.").Build())
			return
		}

		hash, err := auth.HashPassword(body.Password)
		if err != nil {
			log.ErrorContext(ctx, "failed to hash password", logger.Error(err))
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Try again").Build())
			return
		}

		user, err := queries.CreateUser(ctx, models.CreateUserParams{
			Username: body.Username,
			Password: hash,
		})
		if err != nil {
			log.ErrorContext(ctx, "failed to create user", logger.Error(err))
			jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Try again").Build())
		}

		log.DebugContext(ctx, "user created", slog.String("username", body.Username), slog.Int64("user_id", user.ID))
		jw.Write(ctx, w, http.StatusCreated, nil)
	}
}

func LoginHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
	queries *models.Queries,
	store *auth.SessionStore,
) http.HandlerFunc {
	type Request struct {
		Username string `schema:"username" validate:"required,min=3"`
		Password string `schema:"password" validate:"required,min=8,max=1000"`
	}

	typedReader := request.NewTypedReader[Request](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := typedReader.ReadAndValidateForm(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse request body", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		user, err := queries.GetUserByUsername(ctx, body.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.WarnContext(ctx, "Login attempt for non-existent user", slog.String("username", body.Username))
				jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Invalid username or password.").Build())
			} else {
				log.ErrorContext(ctx, "Failed to get user by username during login", slog.String("username", body.Username), logger.Error(err))
				jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Login failed.").Build())
			}
			return
		}

		err = auth.CheckPasswordHash(body.Password, user.Password)
		if err != nil {
			jw.WriteError(ctx, r, w, err)
			return
		}

		err = store.New(w, r, map[string]any{
			"user_id":  user.ID,
			"username": user.Username,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to login user", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		jw.Write(ctx, w, http.StatusOK, nil)
	}
}
