-- name: CreateUser :one
INSERT INTO users (
  username 
) VALUES (
  $1
) RETURNING *;

-- name: ListUsers :many
SELECT * FROM users;

-- name: UpdateUsers :exec
UPDATE users
SET username = $2
WHERE id = $1;
