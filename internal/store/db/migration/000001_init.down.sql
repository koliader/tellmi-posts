-- Drop foreign key constraint first (optional if using CASCADE below, but explicit is safer)
ALTER TABLE comments DROP CONSTRAINT IF EXISTS "comments_post_id_fkey";
ALTER TABLE comments DROP CONSTRAINT IF EXISTS "comments_user_id_fkey";
ALTER TABLE posts DROP CONSTRAINT IF EXISTS "posts_user_id_fkey";
ALTER TABLE posts DROP CONSTRAINT IF EXISTS "posts_category_id_fkey";

-- Drop dependent table first
DROP TABLE IF EXISTS comments;

-- Drop parent table
DROP TABLE IF EXISTS posts;

DROP TABLE IF EXISTS categories;

DROP TABLE IF EXISTS users;
