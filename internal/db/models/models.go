// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package models

import (
	"time"
)

type Card struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Issuer            string   `json:"issuer"`
	Last4Digits       string   `json:"last4_digits"`
	ExpiryDate        string   `json:"expiry_date"`
	DefaultRewardRate *float64 `json:"default_reward_rate"`
	CardType          string   `json:"card_type"`
}

type PredefinedCard struct {
	ID                int64     `json:"id"`
	CardKey           string    `json:"card_key"`
	Name              string    `json:"name"`
	Issuer            string    `json:"issuer"`
	CardType          string    `json:"card_type"`
	DefaultRewardRate float64   `json:"default_reward_rate"`
	RewardType        string    `json:"reward_type"`
	PointValue        float64   `json:"point_value"`
	AnnualFee         int64     `json:"annual_fee"`
	AnnualFeeWaiver   *string   `json:"annual_fee_waiver"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PredefinedRewardRule struct {
	ID               int64     `json:"id"`
	PredefinedCardID int64     `json:"predefined_card_id"`
	Type             string    `json:"type"`
	EntityName       string    `json:"entity_name"`
	RewardRate       float64   `json:"reward_rate"`
	RewardType       string    `json:"reward_type"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
