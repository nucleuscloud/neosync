CREATE TABLE IF NOT EXISTS neosync_api.account_invites (
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
  CONSTRAINT fk_invites_user_id FOREIGN KEY (sender_user_id) REFERENCES neosync_api.users(id) ON DELETE SET NULL
);
