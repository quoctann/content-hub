-- If extensions already exist, move this to public
-- ALTER DATABASE db_name SET search_path TO tenant_a, public;

-- Always set fallback search_path to allow application code use without prefix public
-- Opt 1 - global set search_path: ALTER EXTENSION uuid-ossp SET SCHEMA public;
-- Opt 2 - application code setup whenever open connection: SET search_path = tenant_a, public;

CREATE SCHEMA IF NOT EXISTS public;
CREATE SCHEMA IF NOT EXISTS content;

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