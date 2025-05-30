// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: cards.sql

package models

import (
	"context"
)

const createCard = `-- name: CreateCard :one
INSERT INTO cards (name, -- The name given to the card
                   issuer, -- The issuing institution
                   last4_digits, -- The last four digits of the card number
                   expiry_date, -- The expiration date (e.g., 'MM/YY' or 'YYYY-MM')
                   default_reward_rate, -- The default reward rate (e.g., 1.5 for 1.5%)
                   card_type -- The type of card (e.g., 'Visa', 'Mastercard')
)
VALUES (?, -- Placeholder for Name
        ?, -- Placeholder for Issuer
        ?, -- Placeholder for Last4Digits (ensure the provided value is 4 characters)
        ?, -- Placeholder for ExpiryDate
        ?, -- Placeholder for DefaultRewardRate
        ? -- Placeholder for CardType
       ) RETURNING id, name, issuer, last4_digits, expiry_date, default_reward_rate, card_type
`

type CreateCardParams struct {
	Name              string   `json:"name"`
	Issuer            string   `json:"issuer"`
	Last4Digits       string   `json:"last4_digits"`
	ExpiryDate        string   `json:"expiry_date"`
	DefaultRewardRate *float64 `json:"default_reward_rate"`
	CardType          string   `json:"card_type"`
}

func (q *Queries) CreateCard(ctx context.Context, arg CreateCardParams) (*Card, error) {
	row := q.queryRow(ctx, q.createCardStmt, createCard,
		arg.Name,
		arg.Issuer,
		arg.Last4Digits,
		arg.ExpiryDate,
		arg.DefaultRewardRate,
		arg.CardType,
	)
	var i Card
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Issuer,
		&i.Last4Digits,
		&i.ExpiryDate,
		&i.DefaultRewardRate,
		&i.CardType,
	)
	return &i, err
}

const getAllCards = `-- name: GetAllCards :many
SELECT id, name, issuer, last4_digits, expiry_date, default_reward_rate, card_type FROM cards
ORDER BY name ASC
`

func (q *Queries) GetAllCards(ctx context.Context) ([]*Card, error) {
	rows, err := q.query(ctx, q.getAllCardsStmt, getAllCards)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Card
	for rows.Next() {
		var i Card
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Issuer,
			&i.Last4Digits,
			&i.ExpiryDate,
			&i.DefaultRewardRate,
			&i.CardType,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCardByNameAndIssuer = `-- name: GetCardByNameAndIssuer :one
SELECT id, name, issuer, last4_digits, expiry_date, default_reward_rate, card_type FROM cards
WHERE name = ? AND issuer = ?
LIMIT 1
`

type GetCardByNameAndIssuerParams struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
}

func (q *Queries) GetCardByNameAndIssuer(ctx context.Context, arg GetCardByNameAndIssuerParams) (*Card, error) {
	row := q.queryRow(ctx, q.getCardByNameAndIssuerStmt, getCardByNameAndIssuer, arg.Name, arg.Issuer)
	var i Card
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Issuer,
		&i.Last4Digits,
		&i.ExpiryDate,
		&i.DefaultRewardRate,
		&i.CardType,
	)
	return &i, err
}

const updateCard = `-- name: UpdateCard :one
UPDATE cards
SET name = ?,
    issuer = ?,
    last4_digits = ?,
    expiry_date = ?,
    default_reward_rate = ?,
    card_type = ?
WHERE id = ?
RETURNING id, name, issuer, last4_digits, expiry_date, default_reward_rate, card_type
`

type UpdateCardParams struct {
	Name              string   `json:"name"`
	Issuer            string   `json:"issuer"`
	Last4Digits       string   `json:"last4_digits"`
	ExpiryDate        string   `json:"expiry_date"`
	DefaultRewardRate *float64 `json:"default_reward_rate"`
	CardType          string   `json:"card_type"`
	ID                int64    `json:"id"`
}

func (q *Queries) UpdateCard(ctx context.Context, arg UpdateCardParams) (*Card, error) {
	row := q.queryRow(ctx, q.updateCardStmt, updateCard,
		arg.Name,
		arg.Issuer,
		arg.Last4Digits,
		arg.ExpiryDate,
		arg.DefaultRewardRate,
		arg.CardType,
		arg.ID,
	)
	var i Card
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Issuer,
		&i.Last4Digits,
		&i.ExpiryDate,
		&i.DefaultRewardRate,
		&i.CardType,
	)
	return &i, err
}
