ALTER TABLE neosync_api.account_api_keys
DROP CONSTRAINT account_api_keys_account_id_key_name;

ALTER TABLE neosync_api.account_api_keys
DROP COLUMN key_name;
