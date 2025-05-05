-- name: CreateCard :one
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
       ) RETURNING *;

-- name: GetCardByNameAndIssuer :one
SELECT * FROM cards
WHERE name = ? AND issuer = ?
LIMIT 1;

-- name: GetAllCards :many
SELECT * FROM cards
ORDER BY name ASC;

-- name: UpdateCard :one
UPDATE cards
SET name = ?,
    issuer = ?,
    last4_digits = ?,
    expiry_date = ?,
    default_reward_rate = ?,
    card_type = ?
WHERE id = ?
RETURNING *;