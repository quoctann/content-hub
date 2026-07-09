DROP TRIGGER IF EXISTS trg_account_updated_at ON content.account;

DROP INDEX IF EXISTS idx_account_username;

DROP TABLE IF EXISTS content.account;