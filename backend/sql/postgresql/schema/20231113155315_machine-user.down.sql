ALTER TABLE neosync_api.users
DROP COLUMN IF EXISTS user_type;

ALTER TABLE neosync_api.account_api_keys
DROP COLUMN IF EXISTS user_id;
