DROP FUNCTION IF EXISTS public.trg_update_updated_at();

DROP EXTENSION IF EXISTS "citext" SCHEMA public;
DROP EXTENSION IF EXISTS "uuid-ossp" SCHEMA public;
DROP EXTENSION IF EXISTS "pg_trgm" SCHEMA public;
DROP EXTENSION IF EXISTS "unaccent" SCHEMA public;