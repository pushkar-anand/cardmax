package recommend

import (
	"fmt"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"github.com/pushkar-anand/cardmax/web"
	"log/slog"
	"net/http"
	"time"
)

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
