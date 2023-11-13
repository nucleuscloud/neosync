ALTER TABLE neosync_api.users
DROP COLUMN user_type;

ALTER TABLE neosync_api.account_api_keys
DROP COLUMN user_id;
