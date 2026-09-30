DROP INDEX IF EXISTS idx_content__deleted_at;

ALTER TABLE content DROP COLUMN IF EXISTS deleted_at;
