CREATE TABLE IF NOT EXISTS neosync_api.invites (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
  account_id uuid NOT NULL,
  sender_user_id uuid,
  email varchar NOT NULL,
  token varchar NOT NULL DEFAULT gen_random_uuid(),
  accepted boolean DEFAULT false,
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at timestamp NOT NULL,
	CONSTRAINT invites_pkey PRIMARY KEY (id),
  CONSTRAINT fk_invites_accounts_id FOREIGN KEY (account_id) REFERENCES neosync_api.accounts(id) ON DELETE CASCADE,
  CONSTRAINT fk_invites_users_id FOREIGN KEY (id) REFERENCES neosync_api.users(id) ON DELETE SET NULL
);
ALTER TABLE neosync_api.invites OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.invites TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.invites TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.invites TO neosync_api_readonly;
