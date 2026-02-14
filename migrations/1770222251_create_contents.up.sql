CREATE TABLE IF NOT EXISTS content (
    id BIGSERIAL PRIMARY KEY,
    title TEXT,
    text_data TEXT, -- available for text content
    ocr_text TEXT, -- available for image content after OCR processing
    caption TEXT, -- available for image content
    link TEXT,
    file_name TEXT UNIQUE,
    type TEXT NOT NULL,
    search_vector tsvector,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS content_search_vector_idx ON content USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS content_type_idx ON content (type);
CREATE INDEX IF NOT EXISTS content_created_at_idx ON content (created_at);
CREATE INDEX IF NOT EXISTS content_updated_at_idx ON content (updated_at);

CREATE OR REPLACE FUNCTION content_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', coalesce(NEW.caption, '')), 'A') ||
        setweight(to_tsvector('simple', unaccent(coalesce(NEW.ocr_text, ''))), 'B') ||
        setweight(to_tsvector('simple', coalesce(NEW.text_data, '')), 'C') ||
        setweight(to_tsvector('simple', unaccent(coalesce(NEW.ocr_text, ''))), 'D');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER content_search_vector_update
    BEFORE INSERT OR UPDATE ON content
    FOR EACH ROW
    EXECUTE FUNCTION content_search_vector_update();

CREATE TRIGGER update_content_updated_at
    BEFORE UPDATE ON content
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
