ALTER TABLE neosync_api.account_api_keys
DROP CONSTRAINT IF EXISTS account_api_keys_account_id_key_value;

ALTER TABLE neosync_api.account_api_keys
ADD CONSTRAINT account_api_keys_key_value UNIQUE(key_value);
