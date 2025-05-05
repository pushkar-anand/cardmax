package cards

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	// "time" // No longer needed if getUserIDFromContext is removed

	"github.com/go-chi/chi/v5" // Using chi router, assuming it's used based on common Go practices
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/api/middleware" // Import the middleware package
	"github.com/pushkar-anand/cardmax/internal/db/models" // Use generated models
)

// Remove the placeholder getUserIDFromContext function
/*
func getUserIDFromContext(ctx context.Context) (int64, error) {
	// In a real app, this would extract the user ID set by auth middleware
	userID := ctx.Value("userID") // Assuming "userID" is the context key
	if id, ok := userID.(int64); ok {
		return id, nil
	}
	// Defaulting to user 1 for now, REMOVE THIS IN PRODUCTION
	// return 1, nil
	return 0, errors.New("user ID not found in context or invalid type")
}
*/

// --- User Card Handlers ---

type CardHandler struct {
	Log *slog.Logger
	Jw  *response.JSONWriter
	DB  models.DBTX // Interface for sqlc queries (can be *sql.DB or *sql.Tx)
}

// NewCardHandler creates a new handler instance
func NewCardHandler(log *slog.Logger, jw *response.JSONWriter, db models.DBTX) *CardHandler {
	return &CardHandler{Log: log, Jw: jw, DB: db}
}

// CreateCardRequest defines the expected JSON body for creating a card
type CreateCardRequest struct {
	Name              string   `json:"name"`
	Issuer            string   `json:"issuer"`
	Last4Digits       string   `json:"last4_digits"`
	ExpiryDate        string   `json:"expiry_date"`
	DefaultRewardRate *float64 `json:"default_reward_rate"`
	CardType          string   `json:"card_type"`
}

// CreateCardHandler handles POST requests to create a new user card
func (h *CardHandler) CreateCardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Use the middleware helper function to get user ID
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		// This should ideally not happen if middleware is applied correctly
		h.Log.ErrorContext(ctx, "User ID not found in context after auth middleware")
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Internal server error.").Build())
		return
	}

	var req CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.ErrorContext(ctx, "Failed to decode request body", slog.Any("error", err), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid request body.").Build())
		return
	}
	defer r.Body.Close()

	// TODO: Add input validation for req fields

	queries := models.New(h.DB) // Create sqlc querier

	params := models.CreateCardParams{
		UserID:            userID, // Use userID from context
		Name:              req.Name,
		Issuer:            req.Issuer,
		Last4Digits:       req.Last4Digits,
		ExpiryDate:        req.ExpiryDate,
		DefaultRewardRate: req.DefaultRewardRate,
		CardType:          req.CardType,
	}

	card, err := queries.CreateCard(ctx, params)
	if err != nil {
		h.Log.ErrorContext(ctx, "Failed to create card in DB", slog.Any("error", err), slog.Int64("userID", userID))
		// TODO: Handle specific DB errors (e.g., unique constraint violation)
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to create card.").Build())
		return
	}

	h.Jw.Ok(ctx, w, card)
}

// ListCardsHandler handles GET requests to list cards for the authenticated user
func (h *CardHandler) ListCardsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Use the middleware helper function to get user ID
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		h.Log.ErrorContext(ctx, "User ID not found in context after auth middleware")
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Internal server error.").Build())
		return
	}

	queries := models.New(h.DB)

	cards, err := queries.ListCardsByUser(ctx, userID) // Use userID from context
	if err != nil {
		h.Log.ErrorContext(ctx, "Failed to list cards from DB", slog.Any("error", err), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to retrieve cards.").Build())
		return
	}

	// Return empty list instead of null if no cards found
	if cards == nil {
		cards = []models.Card{}
	}

	h.Jw.Ok(ctx, w, map[string][]models.Card{"cards": cards})
}

// GetCardHandler handles GET requests for a specific card by ID
func (h *CardHandler) GetCardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Use the middleware helper function to get user ID
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		h.Log.ErrorContext(ctx, "User ID not found in context after auth middleware")
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Internal server error.").Build())
		return
	}

	cardIDStr := chi.URLParam(r, "cardID") // Assuming chi router and URL param "cardID"
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		h.Log.WarnContext(ctx, "Invalid card ID format", slog.String("cardID", cardIDStr), slog.Any("error", err), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid card ID format.").Build())
		return
	}

	queries := models.New(h.DB)
	params := models.GetCardByIDAndUserParams{
		ID:     cardID,
		UserID: userID, // Use userID from context
	}

	card, err := queries.GetCardByIDAndUser(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.Log.WarnContext(ctx, "Card not found or access denied", slog.Int64("cardID", cardID), slog.Int64("userID", userID))
			h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusNotFound).WithDetail("Card not found.").Build())
		} else {
			h.Log.ErrorContext(ctx, "Failed to get card from DB", slog.Any("error", err), slog.Int64("cardID", cardID), slog.Int64("userID", userID))
			h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to retrieve card.").Build())
		}
		return
	}

	h.Jw.Ok(ctx, w, card)
}

// UpdateCardRequest defines the expected JSON body for updating a card
type UpdateCardRequest struct {
	Name              string   `json:"name"`
	Issuer            string   `json:"issuer"`
	Last4Digits       string   `json:"last4_digits"`
	ExpiryDate        string   `json:"expiry_date"`
	DefaultRewardRate *float64 `json:"default_reward_rate"`
	CardType          string   `json:"card_type"`
}

// UpdateCardHandler handles PUT/PATCH requests to update a card
func (h *CardHandler) UpdateCardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Use the middleware helper function to get user ID
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		h.Log.ErrorContext(ctx, "User ID not found in context after auth middleware")
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Internal server error.").Build())
		return
	}

	cardIDStr := chi.URLParam(r, "cardID") // Assuming chi router
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		h.Log.WarnContext(ctx, "Invalid card ID format", slog.String("cardID", cardIDStr), slog.Any("error", err), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid card ID format.").Build())
		return
	}

	var req UpdateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log.ErrorContext(ctx, "Failed to decode request body", slog.Any("error", err), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid request body.").Build())
		return
	}
	defer r.Body.Close()

	// TODO: Add input validation

	queries := models.New(h.DB)
	params := models.UpdateCardParams{
		ID:                cardID,
		UserID:            userID, // Use userID from context
		Name:              req.Name,
		Issuer:            req.Issuer,
		Last4Digits:       req.Last4Digits,
		ExpiryDate:        req.ExpiryDate,
		DefaultRewardRate: req.DefaultRewardRate,
		CardType:          req.CardType,
	}

	updatedCard, err := queries.UpdateCard(ctx, params)
	if err != nil {
		// Check if the error is because the card wasn't found for this user
		// sqlc Update returning one might return ErrNoRows if WHERE clause doesn't match
		if errors.Is(err, sql.ErrNoRows) {
			h.Log.WarnContext(ctx, "Card not found for update or access denied", slog.Int64("cardID", cardID), slog.Int64("userID", userID))
			h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusNotFound).WithDetail("Card not found or update failed.").Build())
		} else {
			h.Log.ErrorContext(ctx, "Failed to update card in DB", slog.Any("error", err), slog.Int64("cardID", cardID), slog.Int64("userID", userID))
			h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to update card.").Build())
		}
		return
	}

	h.Jw.Ok(ctx, w, updatedCard)
}

// DeleteCardHandler handles DELETE requests for a specific card by ID
func (h *CardHandler) DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Use the middleware helper function to get user ID
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		h.Log.ErrorContext(ctx, "User ID not found in context after auth middleware")
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Internal server error.").Build())
		return
	}

	cardIDStr := chi.URLParam(r, "cardID") // Assuming chi router
	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		h.Log.WarnContext(ctx, "Invalid card ID format", slog.String("cardID", cardIDStr), slog.Any("error", err), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("Invalid card ID format.").Build())
		return
	}

	queries := models.New(h.DB)
	params := models.DeleteCardParams{
		ID:     cardID,
		UserID: userID, // Use userID from context
	}

	_, err = queries.DeleteCard(ctx, params) // sqlc Delete returns Result and error
	if err != nil {
		// Note: sqlc Delete does not return sql.ErrNoRows if no rows are affected.
		// We might need to check affected rows from the result if precise "not found" is needed.
		// For simplicity, we'll treat errors generally here.
		h.Log.ErrorContext(ctx, "Failed to delete card from DB", slog.Any("error", err), slog.Int64("cardID", cardID), slog.Int64("userID", userID))
		h.Jw.WriteProblem(ctx, r, w, response.NewProblem().WithStatus(http.StatusInternalServerError).WithDetail("Failed to delete card.").Build())
		return
	}

	// Deletion successful, return No Content
	w.WriteHeader(http.StatusNoContent)
}
