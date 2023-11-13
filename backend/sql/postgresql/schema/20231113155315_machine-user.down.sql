ALTER TABLE neosync_api.users
DROP COLUMN user_type;

-- ALTER TABLE neosync_api.account_api_keys
-- DROP CONSTRAINT IF EXISTS fk_account_api_keys_user_id_users_id;
ALTER TABLE neosync_api.account_api_keys
DROP COLUMN user_id;
