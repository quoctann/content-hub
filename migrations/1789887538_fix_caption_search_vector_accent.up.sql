-- caption (weight A) and text_data (weight C) were both indexed with plain
-- to_tsvector(...) — no unaccent() — unlike ocr_text/title which both call
-- unaccent(). Every search query always unaccents its input (see
-- content_repo.go), so any caption/text_data containing accented characters
-- (e.g. Vietnamese diacritics) could never match a search term, because the
-- stored lexeme kept its accent while the query lexeme never did.
CREATE OR REPLACE FUNCTION public.trg_update_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.caption, ''))), 'A') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.ocr_text, ''))), 'B') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.text_data, ''))), 'C') ||
        setweight(to_tsvector('simple', public.unaccent(coalesce(NEW.title, ''))), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SET search_path = public;

-- Backfill: the BEFORE UPDATE trigger recomputes search_vector unconditionally on
-- any UPDATE, so this no-op update forces every existing row to be re-indexed with
-- the corrected function.
UPDATE content SET updated_at = updated_at;
