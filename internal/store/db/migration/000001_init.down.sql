-- Drop foreign key constraint first (optional if using CASCADE below, but explicit is safer)
ALTER TABLE "Comments" DROP CONSTRAINT IF EXISTS "comments_post_id_fkey";

-- Drop dependent table first
DROP TABLE IF EXISTS "Comments";

-- Drop parent table
DROP TABLE IF EXISTS "Posts";
