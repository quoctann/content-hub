DROP INDEX IF NOT EXISTS content_deleted_at_idx;
ALTER TABLE content DROP COLUMN deleted_at;
