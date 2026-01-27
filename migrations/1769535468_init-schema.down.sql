-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop extensions
DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "unaccent";