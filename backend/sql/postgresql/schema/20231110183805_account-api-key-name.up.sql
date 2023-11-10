ALTER TABLE neosync_api.account_api_keys
ADD COLUMN key_name text NOT NULL;

ALTER TABLE neosync_api.account_api_keys
ADD CONSTRAINT account_api_keys_account_id_key_name UNIQUE(account_id, key_name);
