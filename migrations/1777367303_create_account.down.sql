DROP TRIGGER IF EXISTS trg_account_update_updated_at_bu ON content.account;

DROP INDEX IF EXISTS idx_account__username;

DROP TABLE IF EXISTS content.account;