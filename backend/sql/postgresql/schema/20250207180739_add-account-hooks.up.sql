CREATE TABLE IF NOT EXISTS neosync_api.account_hooks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  account_id uuid NOT NULL,

  config jsonb NOT NULL,

  created_by_user_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_by_user_id uuid NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

  enabled boolean NOT NULL DEFAULT true,
  priority integer NOT NULL DEFAULT 0,

  -- hook_timing text GENERATED ALWAYS AS (
  --   CASE
  --     WHEN config->'sql'->'timing'->>'preSync' IS NOT NULL THEN 'preSync'
  --     WHEN config->'sql'->'timing'->>'postSync' IS NOT NULL THEN 'postSync'
  --     ELSE NULL
  --   END
  -- ) STORED,
  -- CONSTRAINT hook_timing_not_null CHECK (hook_timing IS NOT NULL), -- Ensures we always have valid hook timings

  CONSTRAINT fk_account_hooks_account
    FOREIGN KEY (account_id)
    REFERENCES neosync_api.accounts(id)
    ON DELETE CASCADE,

  CONSTRAINT account_hooks_priority_check CHECK (priority >= 0),

  CONSTRAINT account_hooks_name_unique UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS idx_account_hooks_account_id
  ON neosync_api.account_hooks(account_id);

CREATE INDEX IF NOT EXISTS idx_account_hooks_priority
  ON neosync_api.account_hooks(priority);

CREATE INDEX IF NOT EXISTS idx_account_hooks_enabled
  ON neosync_api.account_hooks(enabled)
  WHERE enabled = true;

-- CREATE INDEX IF NOT EXISTS idx_job_hooks_timing_lookup
--   ON neosync_api.job_hooks(job_id, hook_timing, enabled)
--   WHERE enabled = true;

CREATE TRIGGER update_neosync_api_accounthooks_updated_at
  BEFORE UPDATE ON neosync_api.account_hooks
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE neosync_api.account_hooks
  IS 'Stores hooks that can be configured to run as part of an account';
