-- name: CreatePost :one
INSERT INTO "Posts" (
  title,
  description,
  author
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListPosts :many
SELECT * FROM "Posts";

-- name: GetPostByID :one
SELECT * FROM "Posts"
WHERE id = $1;
