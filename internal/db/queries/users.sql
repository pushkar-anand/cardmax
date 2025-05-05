-- name: CreateUser :one
INSERT INTO users (
    username,
    hashed_password
) VALUES (
    ?, ?
)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;
