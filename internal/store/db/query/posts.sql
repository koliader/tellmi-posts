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
-- SELECT * FROM posts;
SELECT p.id, p.title, p.description, u.id as user_id, u.username, c.id as category_id, c.name FROM posts AS p JOIN users AS u ON p.user_id = u.id JOIN categories AS c ON p.category_id = c.id;

-- name: GetPostByID :one
SELECT p.id, p.title, p.description, u.id as user_id, u.username, c.id as category_id, c.name 
FROM posts AS p 
JOIN users AS u ON p.user_id = u.id 
JOIN categories AS c ON p.category_id = c.id 
WHERE p.id = $1;

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
