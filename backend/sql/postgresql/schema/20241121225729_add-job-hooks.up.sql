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
  weight integer NOT NULL DEFAULT 0,

  hook_timing text GENERATED ALWAYS AS (
    CASE
      WHEN config->'sql'->>'pre_sync' IS NOT NULL THEN 'pre_sync'
      WHEN config->'sql'->>'post_sync' IS NOT NULL THEN 'post_sync'
      ELSE NULL
    END
  ) STORED,
  CONSTRAINT hook_timing_not_null CHECK (hook_timing IS NOT NULL), -- Ensures we always have valid hook timings

  CONSTRAINT fk_job_hooks_job
    FOREIGN KEY (job_id)
    REFERENCES neosync_api.jobs(id)
    ON DELETE CASCADE,

  CONSTRAINT job_hooks_weight_check CHECK (weight >= 0),

  CONSTRAINT job_hooks_name_unique UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS idx_job_hooks_job_id
  ON neosync_api.job_hooks(job_id);

CREATE INDEX IF NOT EXISTS idx_job_hooks_weight
  ON neosync_api.job_hooks(weight);

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
