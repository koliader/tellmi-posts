-- name: CreateComment :one
INSERT INTO comments (
  comment,
  post_id,
  user_id
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListCommentsByPost :many
SELECT
    c.id,
    c.comment,
    c.post_id,
    c.user_id,
    cu.username   AS commenter_username,
    p.title       AS post_title,
    p.description AS post_description,
    p.user_id     AS post_author_id,
    pu.username   AS post_author_username,
    cat.id        AS category_id,
    cat.name      AS category_name
FROM comments c
    JOIN posts p       ON p.id = c.post_id
    JOIN users cu      ON cu.id = c.user_id     -- comment author
    JOIN users pu      ON pu.id = p.user_id     -- post author
    JOIN categories cat ON cat.id = p.category_id
WHERE c.post_id = $1
ORDER BY c.id;

-- name: EditComment :one
UPDATE comments
SET comment = $2
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;
