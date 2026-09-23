-- Tables are created unqualified: they land in the schema named by
-- DATABASE_SCHEMA, which the app/migrator send as search_path and create
-- beforehand (database.EnsureSchema). Extensions and shared functions live in
-- public and are always referenced as public.<name>.

CREATE SCHEMA IF NOT EXISTS public;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "citext" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "pg_trgm" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "unaccent" WITH SCHEMA public;

CREATE OR REPLACE FUNCTION public.trg_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql' SET search_path = public; -- lock trigger with public schema