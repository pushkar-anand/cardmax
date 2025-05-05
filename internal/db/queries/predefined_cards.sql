-- name: CreatePredefinedCard :one
INSERT INTO predefined_cards (
    card_key,
    name,
    issuer,
    card_type,
    default_reward_rate,
    reward_type,
    point_value,
    annual_fee,
    annual_fee_waiver
) VALUES (
    ?, -- card_key
    ?, -- name
    ?, -- issuer
    ?, -- card_type
    ?, -- default_reward_rate
    ?, -- reward_type
    ?, -- point_value
    ?, -- annual_fee
    ? -- annual_fee_waiver
)
ON CONFLICT (card_key) DO UPDATE SET
    name = excluded.name,
    issuer = excluded.issuer,
    card_type = excluded.card_type,
    default_reward_rate = excluded.default_reward_rate,
    reward_type = excluded.reward_type,
    point_value = excluded.point_value,
    annual_fee = excluded.annual_fee,
    annual_fee_waiver = excluded.annual_fee_waiver
RETURNING *;

-- name: GetPredefinedCardByKey :one
SELECT * FROM predefined_cards
WHERE card_key = ?
LIMIT 1;

-- name: GetAllPredefinedCards :many
SELECT * FROM predefined_cards
ORDER BY issuer, name;

-- name: CreatePredefinedRewardRule :one
INSERT INTO predefined_reward_rules (
    predefined_card_id,
    type,
    entity_name,
    reward_rate,
    reward_type
) VALUES (
    ?, -- predefined_card_id
    ?, -- type
    ?, -- entity_name
    ?, -- reward_rate
    ? -- reward_type
)
ON CONFLICT (predefined_card_id, type, entity_name) DO UPDATE SET
    reward_rate = excluded.reward_rate,
    reward_type = excluded.reward_type
RETURNING *;

-- name: GetPredefinedRewardRulesByCardID :many
SELECT * FROM predefined_reward_rules
WHERE predefined_card_id = ?
ORDER BY type, entity_name;