CREATE TABLE IF NOT EXISTS neosync_api.account_hooks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  account_id uuid NOT NULL,

  events int[] NOT NULL DEFAULT '{}',

  config jsonb NOT NULL,

  created_by_user_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_by_user_id uuid NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

  enabled boolean NOT NULL DEFAULT true,

  hook_type text GENERATED ALWAYS AS (
    CASE
      WHEN config->'webhook' IS NOT NULL THEN 'webhook'
      ELSE NULL
    END
  ) STORED,
  CONSTRAINT hook_type_not_null CHECK (hook_type IS NOT NULL), -- Ensures we always have a valid hook type

  CONSTRAINT fk_account_hooks_account
    FOREIGN KEY (account_id)
    REFERENCES neosync_api.accounts(id)
    ON DELETE CASCADE,

  CONSTRAINT account_hooks_name_unique UNIQUE (account_id, name)
);

CREATE INDEX idx_account_hooks_account_enabled
  ON neosync_api.account_hooks(account_id, enabled)
  WHERE enabled = true;

-- Create a GIN index for events with the partial condition
CREATE INDEX idx_account_hooks_events_lookup
  ON neosync_api.account_hooks USING GIN (events)
  WHERE enabled = true;

CREATE TRIGGER update_neosync_api_accounthooks_updated_at
  BEFORE UPDATE ON neosync_api.account_hooks
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE neosync_api.account_hooks
  IS 'Stores hooks that can be configured to run as part of an account';
