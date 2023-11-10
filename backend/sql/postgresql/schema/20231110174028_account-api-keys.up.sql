CREATE TABLE IF NOT EXISTS neosync_api.account_api_keys (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
  account_id uuid NOT NULL,

  key_value text NOT NULL,

  created_by_id uuid NOT NULL,
  updated_By_id uuid NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamp NOT NULL,

  CONSTRAINT account_api_keys_pkey PRIMARY KEY (id),
  CONSTRAINT account_api_keys_account_id_key_value UNIQUE(account_id, key_value),
  CONSTRAINT fk_account_api_keys_accounts_id FOREIGN KEY (account_id) REFERENCES neosync_api.accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_account_api_keys_created_by_id FOREIGN KEY (created_by_id) REFERENCES neosync_api.users(id),
  CONSTRAINT fk_account_api_keys_updated_by_id FOREIGN KEY (updated_by_id) REFERENCES neosync_api.users(id)
);
ALTER TABLE neosync_api.account_api_keys OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.account_api_keys TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.account_api_keys TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.account_api_keys TO neosync_api_readonly;
