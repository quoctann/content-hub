ALTER TABLE content.content ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_content__is_hidden ON content.content (is_hidden);