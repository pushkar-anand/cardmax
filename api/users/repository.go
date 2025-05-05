package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/auth"
	"github.com/pushkar-anand/cardmax/internal/db"
	"github.com/pushkar-anand/cardmax/internal/db/models"
	"log/slog"
	"time"
)

// Repository defines the interface for user-related database operations
type Repository interface {
	// CreateUser creates a new user with the given username and password
	CreateUser(ctx context.Context, username, password string) (*User, error)

	// GetUserByUsername retrieves a user by their username
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, id int64) (*User, error)

	// ValidateCredentials validates the username and password combination
	ValidateCredentials(ctx context.Context, username, password string) (*User, error)

	// UserExists checks if a user with the given username exists
	UserExists(ctx context.Context, username string) (bool, error)
}

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// repositoryImpl is a concrete implementation of the Repository interface
type repositoryImpl struct {
	db  *db.DB
	log *slog.Logger
}

// NewRepository creates a new user repository
func NewRepository(db *db.DB, log *slog.Logger) Repository {
	return &repositoryImpl{
		db:  db,
		log: log,
	}
}

// CreateUser creates a new user with the given username and password
func (r *repositoryImpl) CreateUser(ctx context.Context, username, password string) (*User, error) {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to hash password", logger.Error(err))
		return nil, err
	}

	user, err := r.db.Queries.CreateUser(ctx, models.CreateUserParams{
		Username: username,
		Password: hashedPassword,
	})
	if err != nil {
		r.log.ErrorContext(ctx, "failed to create user", slog.Any("error", err), slog.String("username", username))
		return nil, err
	}

	return &User{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

// GetUserByUsername retrieves a user by their username
func (r *repositoryImpl) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := r.db.Queries.GetUserByUsername(ctx, username)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf(`user "%s" not found: %w`, username, err)
	}

	if err != nil {
		r.log.ErrorContext(ctx, "failed to get user by username", slog.Any("error", err), slog.String("username", username))
		return nil, fmt.Errorf(`failed to get user by username: %w`, err)
	}

	u := User{}
	u.FromModel(user)

	return &u, nil
}

// GetUserByID retrieves a user by their ID
func (r *repositoryImpl) GetUserByID(ctx context.Context, id int64) (*User, error) {
	user, err := r.db.Queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		r.log.ErrorContext(ctx, "failed to get user by ID", slog.Any("error", err), slog.Int64("id", id))
		return nil, err
	}

	u := User{}
	u.FromModel(user)

	return &u, nil
}

// ValidateCredentials validates the username and password combination
func (r *repositoryImpl) ValidateCredentials(ctx context.Context, username, password string) (*User, error) {
	user, err := r.db.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrMismatchedHashAndPassword
		}
		r.log.ErrorContext(ctx, "failed to get user during validation", slog.Any("error", err), slog.String("username", username))
		return nil, err
	}

	err = auth.CheckPasswordHash(password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to validate credentials: %w", err)
	}

	u := User{}
	u.FromModel(user)

	return &u, nil
}

// UserExists checks if a user with the given username exists
func (r *repositoryImpl) UserExists(ctx context.Context, username string) (bool, error) {
	_, err := r.db.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.log.ErrorContext(ctx, "failed to check if user exists", slog.Any("error", err), slog.String("username", username))
		return false, err
	}

	return true, nil
}

func (u User) FromModel(m *models.User) *User {
	return &User{
		ID:        m.ID,
		Username:  m.Username,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
