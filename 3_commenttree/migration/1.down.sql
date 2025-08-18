DROP INDEX IF EXISTS idx_comments_parent_id;
DROP INDEX IF EXISTS idx_comments_created_at;
DROP INDEX IF EXISTS idx_comments_content_fts;
DROP TABLE IF EXISTS comments;
DROP EXTENSION IF EXISTS "pgcrypto";