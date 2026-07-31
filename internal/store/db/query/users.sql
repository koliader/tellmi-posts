-- name: CreateUser :one
INSERT INTO users (
  id,
  username 
) VALUES (
  $1, $2
) RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdateUser :one
UPDATE users
SET username = sqlc.arg(new_username)
WHERE id = sqlc.arg(id)
RETURNING *;
