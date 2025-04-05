package models

import (
	"time"
)

// Card represents a credit card in the system
type Card struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Issuer            string    `json:"issuer"`
	Last4Digits       string    `json:"last4Digits"`
	ExpiryDate        time.Time `json:"expiryDate"`
	DefaultRewardRate float64   `json:"defaultRewardRate"`
	CardType          string    `json:"cardType"`
}

// RewardRule represents a specific reward rule for a card
type RewardRule struct {
	ID         int     `json:"id"`
	CardID     int     `json:"cardId"`
	Type       string  `json:"type"` // Category or Merchant
	EntityName string  `json:"entityName"`
	RewardRate float64 `json:"rewardRate"`
	RewardType string  `json:"rewardType"` // Points, Cashback, Miles
	PointValue float64 `json:"pointValue"`
}

// Transaction represents a purchase transaction
type Transaction struct {
	ID           int       `json:"id"`
	Date         time.Time `json:"date"`
	MerchantName string    `json:"merchantName"`
	Category     string    `json:"category"`
	Amount       float64   `json:"amount"`
	CardID       int       `json:"cardId"`
	RewardEarned float64   `json:"rewardEarned"`
	Notes        string    `json:"notes"`
}
