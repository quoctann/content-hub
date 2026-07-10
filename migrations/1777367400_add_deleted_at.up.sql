ALTER TABLE content.content ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_content__deleted_at ON content.content (deleted_at) WHERE deleted_at IS NOT NULL;
