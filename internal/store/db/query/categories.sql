-- name: CreateCategory :one
INSERT INTO categories (
  name
) VALUES (
  $1
) RETURNING *;

-- name: ListCategories :many
SELECT * FROM categories;

-- name: GetCategoryById :one
SELECT * FROM categories
WHERE id = $1;

-- name: EditCategory :exec
UPDATE categories
SET name = $2
WHERE id = $1;
