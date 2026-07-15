-- name: CreateComment :one
INSERT INTO comments (
  comment,
  post_id,
  author
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListCommentsByPost :many
SELECT * FROM comments
WHERE post_id = $1;

-- name: EditComment :one
UPDATE comments
SET comment = $2
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;
