-- name: CreateUser :one
INSERT INTO users (
  username 
) VALUES (
  $1
) RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUser :one
-- name: UpdateUser :one
UPDATE users
SET username = sqlc.arg(new_username)
WHERE username = sqlc.arg(username)
RETURNING *;
