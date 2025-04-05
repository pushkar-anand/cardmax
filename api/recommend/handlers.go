package recommend

import (
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"log/slog"
	"net/http"
	"sort"
)

// GetRecommendationHandler handles recommendation requests
func GetRecommendationHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
) http.HandlerFunc {
	type (
		// RecommendationRequest is the request body for recommendation API
		RecommendationRequest struct {
			Merchant  string  `json:"merchant"`
			Category  string  `json:"category"`
			Amount    float64 `json:"amount"`
			UserCards []int   `json:"user_cards,omitempty"` // Optional: user card IDs to consider
		}

		// RewardResult represents the calculated reward for a card
		RewardResult struct {
			Card        *cards.Card   `json:"card"`
			RewardRate  float64       `json:"reward_rate"`
			RewardType  string        `json:"reward_type"`
			RewardValue float64       `json:"reward_value"`
			CashValue   float64       `json:"cash_value"`
			Rule        *cards.Reward `json:"rule,omitempty"`
		}

		Response struct {
			BestCard *RewardResult   `json:"best_card"`
			AllCards []*RewardResult `json:"all_cards"`
		}
	)

	typedReader := request.NewTypedReader[RecommendationRequest](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := typedReader.ReadAndValidateJSON(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse request body", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		var cardsToUse []*cards.Card
		
		// If user specified cards, only use those
		if len(body.UserCards) > 0 {
			// In a real implementation, we would fetch the user's cards from the database
			// and match them with predefined cards
			// For now, we'll just use all available cards as placeholder
			cardsToUse = cards.GetAll()
		} else {
			// Otherwise, use all available cards
			cardsToUse = cards.GetAll()
		}

		// Calculate rewards for each card
		var results []*RewardResult

		for _, card := range cardsToUse {
			bestRule := findBestRule(body.Merchant, body.Category, card)

			// Calculate reward rate
			rewardRate := card.DefaultRewardRate
			rewardType := card.RewardType

			if bestRule != nil {
				rewardRate = bestRule.RewardRate
				rewardType = bestRule.RewardType
			}

			// Calculate reward value
			rewardValue := (body.Amount * rewardRate) / 100

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

		// Prepare response with best card and all cards
		resp := Response{
			AllCards: results,
		}

		if len(results) > 0 {
			resp.BestCard = results[0]
		}

		jw.Ok(r.Context(), w, resp)
	}
}

// findBestRule finds the best matching rule for a merchant and category
func findBestRule(merchant, category string, card *cards.Card) *cards.Reward {
	var bestRule *cards.Reward
	var bestRate float64 = card.DefaultRewardRate

	for _, rule := range card.RewardRules {
		if (rule.Type == "Merchant" && rule.EntityName == merchant) ||
			(rule.Type == "Category" && rule.EntityName == category) {
			if rule.RewardRate > bestRate {
				bestRule = &rule
				bestRate = rule.RewardRate
			}
		}
	}

	return bestRule
}
