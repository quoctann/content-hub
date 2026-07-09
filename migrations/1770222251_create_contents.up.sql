CREATE TABLE IF NOT EXISTS content.content (
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

CREATE INDEX IF NOT EXISTS idx_content__search_vector ON content.content USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_content__type ON content.content (type);
CREATE INDEX IF NOT EXISTS idx_content__created_at ON content.content (created_at);
CREATE INDEX IF NOT EXISTS idx_content__updated_at ON content.content (updated_at);

CREATE OR REPLACE FUNCTION public.trg_update_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', public.coalesce(NEW.caption, '')), 'A') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.ocr_text, ''))), 'B') ||
        setweight(to_tsvector('simple', public.coalesce(NEW.text_data, '')), 'C') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.ocr_text, ''))), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SET search_path = public;

CREATE TRIGGER trg_content_update_search_vector_biu
    BEFORE INSERT OR UPDATE ON content.content
    FOR EACH ROW
    EXECUTE FUNCTION public.trg_update_search_vector()

CREATE TRIGGER trg_content_update_updated_at_bu
    BEFORE UPDATE ON content.content
    FOR EACH ROW
    EXECUTE FUNCTION public.shared__all__fn_trig_update_updated_at();
