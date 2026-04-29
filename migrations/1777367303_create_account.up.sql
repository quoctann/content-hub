CREATE TABLE IF NOT EXISTS account (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_account_username ON account(username);

DROP TRIGGER IF EXISTS trg_account_updated_at ON account;

CREATE TRIGGER trg_account_updated_at
    BEFORE UPDATE ON account
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
