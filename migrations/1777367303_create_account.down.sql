DROP TRIGGER IF EXISTS trg_account_updated_at ON account;

DROP INDEX IF EXISTS idx_account_username;

DROP TABLE IF EXISTS account;