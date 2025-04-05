package recommend

import (
	"encoding/json"
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"log/slog"
	"net/http"
	"sort"
)

// RecommendationRequest is the request body for recommendation API
type RecommendationRequest struct {
	Merchant string  `json:"merchant"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	UserCards []int  `json:"user_cards,omitempty"` // Optional: user card IDs to consider
}

// RewardResult represents calculated reward for a card
type RewardResult struct {
	Card        *cards.Card `json:"card"`
	RewardRate  float64     `json:"reward_rate"`
	RewardType  string      `json:"reward_type"`
	RewardValue float64     `json:"reward_value"`
	CashValue   float64     `json:"cash_value"`
	Rule        *cards.Reward `json:"rule,omitempty"`
}

// RecommendationResponse is the response for recommendation API
type RecommendationResponse struct {
	BestCard *RewardResult   `json:"best_card"`
	AllCards []*RewardResult `json:"all_cards"`
}

// GetRecommendationHandler handles recommendation requests
func GetRecommendationHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RecommendationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jw.WriteProblem(r.Context(), r, w, response.NewProblem().WithStatus(http.StatusBadRequest).WithDetail("invalid request").WithParam("error", err.Error()).Build())
			return
		}

		// Get all available cards
		allCards := cards.GetAll()
		
		// Calculate rewards for each card
		var results []*RewardResult
		
		for _, card := range allCards {
			bestRule := findBestRule(req.Merchant, req.Category, card)
			
			// Calculate reward rate
			rewardRate := card.DefaultRewardRate
			rewardType := card.RewardType
			
			if bestRule != nil {
				rewardRate = bestRule.RewardRate
				rewardType = bestRule.RewardType
			}
			
			// Calculate reward value
			rewardValue := (req.Amount * rewardRate) / 100
			
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
		response := RecommendationResponse{
			AllCards: results,
		}
		
		if len(results) > 0 {
			response.BestCard = results[0]
		}
		
		jw.Ok(r.Context(), w, response)
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