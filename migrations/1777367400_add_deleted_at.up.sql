ALTER TABLE content ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS content_deleted_at_idx ON content (deleted_at);
