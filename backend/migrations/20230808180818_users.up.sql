CREATE TABLE IF NOT EXISTS neosync_api.users (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);
ALTER TABLE neosync_api.users OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.users TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.users TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.users TO neosync_api_readonly;

CREATE TABLE IF NOT EXISTS neosync_api.user_identity_provider_associations (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	user_id uuid NOT NULL,
	auth0_provider_id varchar NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT now(),
	CONSTRAINT user_identity_provider_associations_auth0_provider_id_key UNIQUE (auth0_provider_id),
	CONSTRAINT user_identity_provider_associations_pkey PRIMARY KEY (id),
	CONSTRAINT fk_user_identity_provider_user_id FOREIGN KEY (user_id) REFERENCES neosync_api.users(id) ON DELETE CASCADE
);
ALTER TABLE neosync_api.user_identity_provider_associations OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.user_identity_provider_associations TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.user_identity_provider_associations TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.user_identity_provider_associations TO neosync_api_readonly;

CREATE TABLE IF NOT EXISTS neosync_api.accounts (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  account_type smallint NOT NULL,
	account_slug varchar NOT NULL,
	-- subscription_status text NULL,
	-- trial_expires_at timestamp NULL,
	-- subscription_expires_at timestamp NULL,
	-- trial_environment_limit int2 NULL,
	CONSTRAINT accounts_pkey PRIMARY KEY (id)
);
ALTER TABLE neosync_api.accounts OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.accounts TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.accounts TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.accounts TO neosync_api_readonly;

CREATE TABLE IF NOT EXISTS neosync_api.account_user_associations (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	account_id uuid NOT NULL,
	user_id uuid NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT account_user_associations_account_id_user_id_key UNIQUE (account_id, user_id),
	CONSTRAINT account_user_associations_pkey PRIMARY KEY (id),
	CONSTRAINT fk_account_user_associations_account_id FOREIGN KEY (account_id) REFERENCES neosync_api.accounts(id) ON DELETE CASCADE,
	CONSTRAINT fk_account_user_associations_user_id FOREIGN KEY (user_id) REFERENCES neosync_api.users(id) ON DELETE CASCADE
);
ALTER TABLE neosync_api.account_user_associations OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.account_user_associations TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.account_user_associations TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.account_user_associations TO neosync_api_readonly;
