-- Enable UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable CITEXT for case-insensitive text fields (e.g. email)
CREATE EXTENSION IF NOT EXISTS "citext";

-- Create a function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Search feature
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";
-- Function chuẩn hóa tiếng Việt
CREATE OR REPLACE FUNCTION normalize_vietnamese(text)
RETURNS text AS $$
  SELECT lower(unaccent($1))
$$ LANGUAGE sql IMMUTABLE;
