CREATE TABLE IF NOT EXISTS neosync_api.job_hooks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  job_id uuid NOT NULL,

  config jsonb NOT NULL,

  created_by_user_id uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_by_user_id uuid NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

  enabled boolean NOT NULL DEFAULT true,
  priority integer NOT NULL DEFAULT 0,

  hook_timing text GENERATED ALWAYS AS (
    CASE
      WHEN config->'sql'->'timing'->>'preSync' IS NOT NULL THEN 'preSync'
      WHEN config->'sql'->'timing'->>'postSync' IS NOT NULL THEN 'postSync'
      ELSE NULL
    END
  ) STORED,
  CONSTRAINT hook_timing_not_null CHECK (hook_timing IS NOT NULL), -- Ensures we always have valid hook timings

  connection_id uuid GENERATED ALWAYS AS (
    CASE
      WHEN config ? 'sql' THEN (config->'sql'->>'connectionId')::uuid
      ELSE NULL
    END
  ) STORED,

  CONSTRAINT fk_job_hooks_connection
    FOREIGN KEY (connection_id)
    REFERENCES neosync_api.connections(id)
    ON DELETE RESTRICT,

  CONSTRAINT fk_job_hooks_job
    FOREIGN KEY (job_id)
    REFERENCES neosync_api.jobs(id)
    ON DELETE CASCADE,

  CONSTRAINT job_hooks_priority_check CHECK (priority >= 0),

  CONSTRAINT job_hooks_name_unique UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS idx_job_hooks_job_id
  ON neosync_api.job_hooks(job_id);

CREATE INDEX IF NOT EXISTS idx_job_hooks_priority
  ON neosync_api.job_hooks(priority);

CREATE INDEX IF NOT EXISTS idx_job_hooks_enabled
  ON neosync_api.job_hooks(enabled)
  WHERE enabled = true;

CREATE INDEX IF NOT EXISTS idx_job_hooks_timing_lookup
  ON neosync_api.job_hooks(job_id, hook_timing, enabled)
  WHERE enabled = true;

CREATE TRIGGER update_neosync_api_jobhooks_updated_at
  BEFORE UPDATE ON neosync_api.job_hooks
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE neosync_api.job_hooks
  IS 'Stores hooks that can be configured to run as part of a job';
