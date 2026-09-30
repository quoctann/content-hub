CREATE OR REPLACE FUNCTION public.trg_update_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.caption, '')), 'A') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.ocr_text, ''))), 'B') ||
        setweight(to_tsvector('simple', coalesce(NEW.text_data, '')), 'C') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.title, ''))), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SET search_path = public;

UPDATE content SET updated_at = updated_at;
