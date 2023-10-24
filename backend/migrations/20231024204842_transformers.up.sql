CREATE TABLE IF NOT EXISTS neosync_api.transformers (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  name varchar NOT NULL,
  account_id uuid NOT NULL,
  transformer_config jsonb NOT NULL,

  created_by_id uuid NOT NULL,
  updated_by_id uuid NOT NULL,

  CONSTRAINT transformers_pkey PRIMARY KEY (id),
  CONSTRAINT fk_transformers_accounts_id FOREIGN KEY (account_id) REFERENCES neosync_api.accounts(id) ON DELETE CASCADE,
  CONSTRAINT transformers_name_account_id UNIQUE (name, account_id),
  CONSTRAINT fk_transformers_created_by_users_id FOREIGN KEY (created_by_id) REFERENCES neosync_api.users(id),
  CONSTRAINT fk_transformers_updated_by_users_id FOREIGN KEY (updated_by_id) REFERENCES neosync_api.users(id)
);
ALTER TABLE neosync_api.transformers OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.transformers TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.transformers TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.transformers TO neosync_api_readonly;