package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/sessions" // Import sessions
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/internal/auth" // Import the auth package
	"github.com/pushkar-anand/cardmax/internal/db/models" // Import the generated models
)

// UserHandler holds dependencies for user-related handlers.
type UserHandler struct {
	Log   *slog.Logger
	Jw    *response.JSONWriter
	DB    models.DBTX // Database interface (sql.DB or sql.Tx)
	Store sessions.Store // Add session store
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(log *slog.Logger, jw *response.JSONWriter, db models.DBTX, store sessions.Store) *UserHandler {
	return &UserHandler{Log: log, Jw: jw, DB: db, Store: store} // Include store in initialization
}

// RegisterRequest defines the expected JSON body for user registration.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse defines the JSON body returned upon successful registration.
type RegisterResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// LoginRequest defines the expected JSON body for user login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines the JSON body returned upon successful login.
type LoginResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// RegisterUserHandler handles POST requests for user registration.
func (h *UserHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Decode Request Body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.ErrorContext(ctx, "Failed to decode registration request body", slog.Any("error", err))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid request body.").Build())
		return
	}
	defer r.Body.Close()

	// 2. Basic Validation
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password) // Trim password too? Usually no, but for consistency here.

	if req.Username == "" {
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Username cannot be empty.").Build())
		return
	}
	if len(req.Password) < 8 { // Example: Minimum password length
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Password must be at least 8 characters long.").Build())
		return
	}

	queries := models.New(h.DB) // Create sqlc querier

	// 3. Check if Username Exists
	_, err := queries.GetUserByUsername(ctx, req.Username)
	if err == nil {
		// User found, username already exists
		h.Log.WarnContext(ctx, "Registration attempt with existing username", slog.String("username", req.Username))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusConflict).WithDetail("Username already taken.").Build())
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		// An unexpected database error occurred
		h.Log.ErrorContext(ctx, "Failed to check for existing username", slog.String("username", req.Username), slog.Any("error", err))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to process registration.").Build())
		return
	}
	// If err is sql.ErrNoRows, proceed with registration

	// 4. Hash Password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		h.Log.ErrorContext(ctx, "Failed to hash password during registration", slog.String("username", req.Username), slog.Any("error", err))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to process registration.").Build())
		return
	}

	// 5. Create User in DB
	params := models.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
	}
	newUser, err := queries.CreateUser(ctx, params)
	if err != nil {
		h.Log.ErrorContext(ctx, "Failed to create user in database", slog.String("username", req.Username), slog.Any("error", err))
		// TODO: Handle specific DB errors like unique constraint violation if possible (though username check should prevent it)
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to complete registration.").Build())
		return
	}

	// 6. Success Response
	h.Log.InfoContext(ctx, "User registered successfully", slog.String("username", newUser.Username), slog.Int64("userID", newUser.ID))
	resp := RegisterResponse{
		ID:       newUser.ID,
		Username: newUser.Username,
	}
	h.Jw.Write(ctx, w, http.StatusCreated, resp) // Use Write for 201 Created
}

// LoginUserHandler handles POST requests for user login.
func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Decode Request Body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.ErrorContext(ctx, "Failed to decode login request body", slog.Any("error", err))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid request body.").Build())
		return
	}
	defer r.Body.Close()

	// 2. Basic Validation
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Username and password cannot be empty.").Build())
		return
	}

	queries := models.New(h.DB)

	// 3. Get User By Username
	user, err := queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.Log.WarnContext(ctx, "Login attempt for non-existent user", slog.String("username", req.Username))
			h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Invalid username or password.").Build())
		} else {
			h.Log.ErrorContext(ctx, "Failed to get user by username during login", slog.String("username", req.Username), slog.Any("error", err))
			h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Login failed.").Build())
		}
		return
	}

	// 4. Check Password Hash
	if !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		h.Log.WarnContext(ctx, "Incorrect password attempt", slog.String("username", req.Username), slog.Int64("userID", user.ID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusUnauthorized).WithDetail("Invalid username or password.").Build())
		return
	}

	// 5. Get Session
	// Use the session store injected into the handler
	session, err := h.Store.Get(r, "user-session") // Use a consistent session name
	if err != nil {
		// Note: Gorilla sessions might return an error even if the session is just new.
		// It's usually safe to proceed and set values unless the error is critical.
		// However, logging it is good practice.
		h.Log.WarnContext(ctx, "Error getting session (might be new)", slog.String("username", user.Username), slog.Int64("userID", user.ID), slog.Any("error", err))
		// Don't return here unless err indicates a fundamental store issue.
	}

	// 6. Set Session Values
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username // Optional: store username too

	// 7. Set Session Options (Example)
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		// Secure: true, // Enable this if using HTTPS
		// SameSite: http.SameSiteLaxMode, // Good default for CSRF protection
	}

	// 8. Save Session
	err = session.Save(r, w)
	if err != nil {
		h.Log.ErrorContext(ctx, "Failed to save session", slog.String("username", user.Username), slog.Int64("userID", user.ID), slog.Any("error", err))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Login failed (session error).").Build())
		return
	}

	// 9. Success Response
	h.Log.InfoContext(ctx, "User logged in successfully", slog.String("username", user.Username), slog.Int64("userID", user.ID))
	resp := LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Message:  "Login successful",
	}
	h.Jw.Ok(ctx, w, resp) // Use Ok for 200 OK
}
