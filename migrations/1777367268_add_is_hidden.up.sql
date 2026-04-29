ALTER TABLE content ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS content_is_hidden_idx ON content (is_hidden);