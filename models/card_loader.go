package models

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// PredefinedCard extends the Card struct with additional information
type PredefinedCard struct {
	Card
	RewardType      string       `json:"rewardType"`
	PointValue      float64      `json:"pointValue"`
	AnnualFee       float64      `json:"annualFee"`
	AnnualFeeWaiver string       `json:"annualFeeWaiver"`
	RewardRules     []RewardRule `json:"rewardRules"`
	Benefits        []string     `json:"benefits"`
}

// LoadPredefinedCards loads all predefined card definitions from the data/cards directory
func LoadPredefinedCards() ([]PredefinedCard, error) {
	var cards []PredefinedCard
	
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	
	// Path to the cards directory
	cardsDir := filepath.Join(cwd, "data", "cards")
	
	// Read all files in the directory
	files, err := ioutil.ReadDir(cardsDir)
	if err != nil {
		return nil, err
	}
	
	// Process each JSON file
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(cardsDir, file.Name())
			
			// Read the file
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			
			// Parse the JSON
			var card PredefinedCard
			if err := json.Unmarshal(data, &card); err != nil {
				return nil, err
			}
			
			cards = append(cards, card)
		}
	}
	
	return cards, nil
}

// GetPredefinedCardByID returns a predefined card by its ID
func GetPredefinedCardByID(id int) (PredefinedCard, error) {
	cards, err := LoadPredefinedCards()
	if err != nil {
		return PredefinedCard{}, err
	}
	
	for _, card := range cards {
		if card.ID == id {
			return card, nil
		}
	}
	
	return PredefinedCard{}, os.ErrNotExist
}

// GetPredefinedCardByName returns a predefined card by its name
func GetPredefinedCardByName(name string) (PredefinedCard, error) {
	cards, err := LoadPredefinedCards()
	if err != nil {
		return PredefinedCard{}, err
	}
	
	for _, card := range cards {
		if strings.EqualFold(card.Name, name) {
			return card, nil
		}
	}
	
	return PredefinedCard{}, os.ErrNotExist
}

// ConvertToCard converts a PredefinedCard to a regular Card
func (pc *PredefinedCard) ConvertToCard() Card {
	return Card{
		ID:                pc.ID,
		Name:              pc.Name,
		Issuer:            pc.Issuer,
		Last4Digits:       "", // This needs to be provided by the user
		ExpiryDate:        pc.ExpiryDate,
		DefaultRewardRate: pc.DefaultRewardRate,
		CardType:          pc.CardType,
	}
}

// GetRewardRules returns the reward rules for a predefined card
func (pc *PredefinedCard) GetRewardRules() []RewardRule {
	// Make a copy of the rules to avoid modifying the original
	rules := make([]RewardRule, len(pc.RewardRules))
	copy(rules, pc.RewardRules)
	
	// Set the CardID for each rule
	for i := range rules {
		rules[i].CardID = pc.ID
	}
	
	return rules
}
