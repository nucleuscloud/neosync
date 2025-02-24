CREATE TABLE IF NOT EXISTS neosync_api.slack_oauth_connections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  account_id uuid NOT NULL,
  access_token text NOT NULL,

  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_slack_oauth_connections_account
    FOREIGN KEY (account_id)
    REFERENCES neosync_api.accounts(id)
    ON DELETE CASCADE,

  CONSTRAINT slack_oauth_connections_account_id_access_token_unique UNIQUE (account_id, access_token)
);

CREATE INDEX IF NOT EXISTS idx_slack_oauth_connections_account_id
  ON neosync_api.slack_oauth_connections(account_id);

CREATE TRIGGER update_neosync_api_slack_oauth_connections_updated_at
  BEFORE UPDATE ON neosync_api.slack_oauth_connections
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE neosync_api.slack_oauth_connections
  IS 'Stores Slack OAuth connections for a given account';

ALTER TABLE neosync_api.account_hooks
  ALTER COLUMN hook_type SET GENERATED ALWAYS AS (
    CASE
      WHEN config->'webhook' IS NOT NULL THEN 'webhook'
      WHEN config->'slack' IS NOT NULL THEN 'slack'
      ELSE NULL
    END
  ) STORED;
