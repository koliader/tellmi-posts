-- name: CreatePost :one
INSERT INTO posts (
  title,
  description,
  author
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListPosts :many
SELECT * FROM posts;

-- name: GetPostByID :one
SELECT * FROM posts
WHERE id = $1;

-- name: EditPost :one
UPDATE posts
SET title = $2,
description = $3
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;
