ALTER TABLE neosync_api.account_api_keys
DROP CONSTRAINT IF EXISTS account_api_keys_account_id_key_name;

ALTER TABLE neosync_api.account_api_keys
DROP COLUMN IF EXISTS key_name;
