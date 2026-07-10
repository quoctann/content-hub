DROP INDEX IF NOT EXISTS idx_content__deleted_at;

ALTER TABLE content.content DROP COLUMN IF EXISTS deleted_at;
