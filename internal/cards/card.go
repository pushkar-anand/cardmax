package cards

import (
	"embed"
	"encoding/json"
	"fmt"
)

type (
	// Reward represents rewards on a card
	Reward struct {
		Type       string  `json:"type"`
		EntityName string  `json:"entity_name"`
		RewardRate float64 `json:"reward_rate"`
		RewardType string  `json:"reward_type"`
	}

	// Card represents a credit card in the system
	Card struct {
		Key               string   `json:"card_key"`
		Name              string   `json:"name"`
		Issuer            string   `json:"issuer"`
		CardType          string   `json:"card_type"`
		DefaultRewardRate float64  `json:"default_reward_rate"`
		RewardType        string   `json:"reward_type"`
		PointValue        float64  `json:"point_value"`
		AnnualFee         int      `json:"annual_fee"`
		AnnualFeeWaiver   string   `json:"annual_fee_waiver"`
		RewardRules       []Reward `json:"reward_rules"`
		Benefits          []string `json:"benefits"`
	}

	parsedCards struct {
		cards   []*Card
		cardMap map[string]*Card
	}
)

var parsed parsedCards

// Parse parses all predefined card definitions from the data/cards directory
func Parse(data embed.FS) error {
	entries, err := data.ReadDir("cards")
	if err != nil {
		return fmt.Errorf("error reading cards dir: %w", err)
	}

	cards := make([]*Card, 0, len(entries))
	cardMap := make(map[string]*Card)

	for _, entry := range entries {
		fn := fmt.Sprintf("cards/%s", entry.Name())

		file, err := data.ReadFile(fn)
		if err != nil {
			return fmt.Errorf("error reading card file %s: %w", fn, err)
		}

		var card Card

		err = json.Unmarshal(file, &card)
		if err != nil {
			return fmt.Errorf("error unmarshalling card data %s: %w", fn, err)
		}

		cards = append(cards, &card)
		cardMap[card.Key] = &card
	}

	parsed.cards = cards
	parsed.cardMap = cardMap

	return nil
}

// GetAll returns all predefined cards
func GetAll() []*Card {
	return parsed.cards
}

// GetByKey returns a card by its key
func GetByKey(key string) (*Card, bool) {
	card, ok := parsed.cardMap[key]
	return card, ok
}
