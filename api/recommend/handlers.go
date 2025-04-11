package recommend

import (
	"github.com/pushkar-anand/build-with-go/http/request"
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"github.com/pushkar-anand/cardmax/web"
	"log/slog"
	"net/http"
	"sort"
)

type (
	// RecommendationRequest is the request body for recommendation API
	RecommendationRequest struct {
		Merchant string  `json:"merchant" validate:"required_without=Category"`
		Category string  `json:"category" validate:"required_without=Merchant"`
		Amount   float64 `json:"amount" validate:"required,min=1"`
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
)

// GetRecommendationHandler handles recommendation requests
func GetRecommendationHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
	reader *request.Reader,
) http.HandlerFunc {
	type (
		Request struct {
			RecommendationRequest
		}

		Response struct {
			BestCard *RewardResult   `json:"best_card"`
			AllCards []*RewardResult `json:"all_cards"`
		}
	)

	typedReader := request.NewTypedReader[Request](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := typedReader.ReadAndValidateJSON(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse request body", logger.Error(err))
			jw.WriteError(ctx, r, w, err)
			return
		}

		var cardsToUse []*cards.Card

		cardsToUse = cards.GetAll()

		best, all := analyzeCards(cardsToUse, body.RecommendationRequest)

		// Prepare a response with the best card and all cards
		resp := Response{
			BestCard: best,
			AllCards: all,
		}

		jw.Ok(r.Context(), w, resp)
	}
}

// GetRecommendationHTMLHandler handles recommendation requests and returns HTML
func GetRecommendationHTMLHandler(
	log *slog.Logger,
	reader *request.Reader,
	tr *web.Renderer,
) http.HandlerFunc {
	type (
		Request struct {
			RecommendationRequest
		}
	)

	typedReader := request.NewTypedReader[Request](reader)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data, err := typedReader.ReadAndValidateForm(r)
		if err != nil {
			log.ErrorContext(ctx, "failed to parse form", logger.Error(err))
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// For now, use all available cards (later can be based on user selection)
		cardsToUse := cards.GetAll()

		best, all := analyzeCards(cardsToUse, data.RecommendationRequest)

		// Prepare template data
		tmplData := map[string]interface{}{
			"AllCards": all,
			"Amount":   data.Amount,
			"Merchant": data.Merchant,
			"Category": data.Category,
			"BestCard": best,
		}

		// Add the best card if there are results
		if best != nil {
			tmplData["BestCard"] = best
		}

		// Render the HTML template
		err = tr.RenderPartial(w, web.PartialRecommendationResult, tmplData)
		if err != nil {
			log.ErrorContext(ctx, "error rendering recommendation template", logger.Error(err))
			http.Error(w, "Failed to render recommendation", http.StatusInternalServerError)
			return
		}
	}
}

func analyzeCards(cardsToUse []*cards.Card, rr RecommendationRequest) (best *RewardResult, all []*RewardResult) {
	all = make([]*RewardResult, 0, len(cardsToUse))

	for _, card := range cardsToUse {
		bestRule := findBestRule(rr.Merchant, rr.Category, card)

		// Calculate reward rate
		rewardRate := card.DefaultRewardRate
		rewardType := card.RewardType

		if bestRule != nil {
			rewardRate = bestRule.RewardRate
			rewardType = bestRule.RewardType
		}

		// Calculate reward value
		rewardValue := (rr.Amount * rewardRate) / 100

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

		all = append(all, result)
	}

	// Sort by cash value (highest first)
	sort.Slice(all, func(i, j int) bool {
		return all[i].CashValue > all[j].CashValue
	})

	if len(all) == 0 {
		return nil, all
	}

	return all[0], all
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
