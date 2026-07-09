DROP INDEX IF NOT EXISTS idx_content__deleted_at;

ALTER TABLE content.content DROP COLUMN deleted_at;
