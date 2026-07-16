-- name: CreatePost :one
INSERT INTO posts (
  title,
  description,
  user_id,
  category_id
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: ListPosts :many
SELECT * FROM posts;

-- name: GetPostByID :one
SELECT * FROM posts
WHERE id = $1;

-- name: EditPost :one
UPDATE posts
SET title = $2,
description = $3,
category_id = $4
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;
