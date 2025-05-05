-- name: CreateCard :one
INSERT INTO cards (name,
                   issuer,
                   last4_digits,
                   expiry_date,
                   default_reward_rate,
                   card_type,
                   user_id -- Added user_id
)
VALUES (?, ?, ?, ?, ?, ?, ?) -- Added placeholder for user_id
RETURNING *;

-- name: GetCardByNameIssuerAndUser :one
SELECT * FROM cards
WHERE name = ? AND issuer = ? AND user_id = ? -- Added user_id filter
LIMIT 1;

-- name: ListCardsByUser :many
SELECT * FROM cards
WHERE user_id = ? -- Added user_id filter
ORDER BY name ASC;

-- name: UpdateCard :one
UPDATE cards
SET name = ?,
    issuer = ?,
    last4_digits = ?,
    expiry_date = ?,
    default_reward_rate = ?,
    card_type = ?
WHERE id = ? AND user_id = ? -- Added user_id filter
RETURNING *;

-- Add a DeleteCard query as well, ensuring user_id check
-- name: DeleteCard :exec
DELETE FROM cards
WHERE id = ? AND user_id = ?;

-- name: GetCardByIDAndUser :one
SELECT * FROM cards
WHERE id = ? AND user_id = ?;