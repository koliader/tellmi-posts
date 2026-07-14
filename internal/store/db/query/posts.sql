-- name: CreatePost :one
INSERT INTO "Posts" (
  title,
  description,
  author
) VALUES (
  $1, $2, $3
) RETURNING *;
