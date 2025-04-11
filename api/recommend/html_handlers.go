package recommend

import (
	"fmt"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"github.com/pushkar-anand/cardmax/web"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"
)

// GetRecommendationHTMLHandler handles recommendation requests and returns HTML
func GetRecommendationHTMLHandler(
	log *slog.Logger,
	reader *request.Reader,
	tr *web.Renderer,
) http.HandlerFunc {
	type (
		// RecommendationRequest is the request body for recommendation API
		RecommendationRequest struct {
			Merchant  string  `json:"merchant"`
			Category  string  `json:"category"`
			Amount    float64 `json:"amount"`
			UserCards []int   `json:"user_cards,omitempty"` // Optional: user card IDs to consider
		}
	)

	typedReader := request.NewTypedReader[RecommendationRequest](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse form data
		if err := r.ParseForm(); err != nil {
			log.ErrorContext(ctx, "failed to parse form", logger.Error(err))
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Extract form values
		merchant := strings.TrimSpace(r.FormValue("merchant"))
		category := strings.TrimSpace(r.FormValue("category"))
		amountStr := r.FormValue("amount")

		// Validation
		if merchant == "" && category == "" {
			log.ErrorContext(ctx, "both merchant and category are empty")
			http.Error(w, "Please provide either merchant or category", http.StatusBadRequest)
			return
		}

		// Parse amount
		var amount float64
		if amountStr != "" {
			if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
				log.ErrorContext(ctx, "invalid amount format", logger.Error(err))
				http.Error(w, "Invalid amount format", http.StatusBadRequest)
				return
			}
		}

		// For now, use all available cards (later can be based on user selection)
		cardsToUse := cards.GetAll()

		// Calculate rewards for each card
		var results []*RewardResult

		for _, card := range cardsToUse {
			bestRule := findBestRule(merchant, category, card)

			// Calculate reward rate
			rewardRate := card.DefaultRewardRate
			rewardType := card.RewardType

			if bestRule != nil {
				rewardRate = bestRule.RewardRate
				rewardType = bestRule.RewardType
			}

			// Calculate reward value
			rewardValue := (amount * rewardRate) / 100

			// Calculate cash value
			cashValue := rewardValue
			if rewardType == "Points" || rewardType == "Miles" {
				cashValue = rewardValue * card.PointValue
			}

			result := &RewardResult{
				Card:        card,
				RewardRate:  rewardRate,
				RewardType:  rewardType,
				RewardValue: rewardValue,
				CashValue:   cashValue,
				Rule:        bestRule,
			}

			results = append(results, result)
		}

		// Sort by cash value (highest first)
		sort.Slice(results, func(i, j int) bool {
			return results[i].CashValue > results[j].CashValue
		})

		// Prepare template data
		data := map[string]interface{}{
			"AllCards": results,
			"Amount":   amount,
			"Merchant": merchant,
			"Category": category,
		}

		// Add best card if there are results
		if len(results) > 0 {
			data["BestCard"] = results[0]
		}

		// Render the HTML template
		err := tr.RenderPartial(w, "recommendation_result", data)
		if err != nil {
			log.ErrorContext(ctx, "error rendering recommendation template", logger.Error(err))
			http.Error(w, "Failed to render recommendation", http.StatusInternalServerError)
			return
		}
	}
}

// GetSaveTransactionFormHandler returns the save transaction form with pre-filled data
func GetSaveTransactionFormHandler(
	log *slog.Logger,
	tr *web.Renderer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse query params
		merchant := r.URL.Query().Get("merchant")
		category := r.URL.Query().Get("category")
		amountStr := r.URL.Query().Get("amount")

		// Parse amount
		var amount float64
		if amountStr != "" {
			if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
				log.ErrorContext(ctx, "invalid amount format", logger.Error(err))
				http.Error(w, "Invalid amount format", http.StatusBadRequest)
				return
			}
		}

		// Calculate reward based on best card
		var rewardEarned float64
		allCards := cards.GetAll()
		if len(allCards) > 0 {
			bestCard := allCards[0]
			bestRule := findBestRule(merchant, category, bestCard)

			rewardRate := bestCard.DefaultRewardRate
			if bestRule != nil {
				rewardRate = bestRule.RewardRate
			}

			rewardValue := (amount * rewardRate) / 100
			rewardEarned = rewardValue
			if bestCard.RewardType == "Points" || bestCard.RewardType == "Miles" {
				rewardEarned = rewardValue * bestCard.PointValue
			}
		}

		// Prepare data for template
		// In a real app, this would include user's cards
		data := map[string]interface{}{
			"Merchant":     merchant,
			"Category":     category,
			"Amount":       amount,
			"Date":         time.Now().Format("2006-01-02"),
			"Cards":        allCards,
			"RewardEarned": rewardEarned,
		}

		err := tr.RenderPartial(w, "save_transaction", data)
		if err != nil {
			log.ErrorContext(ctx, "error rendering save transaction template", logger.Error(err))
			http.Error(w, "Failed to render save transaction form", http.StatusInternalServerError)
			return
		}
	}
}