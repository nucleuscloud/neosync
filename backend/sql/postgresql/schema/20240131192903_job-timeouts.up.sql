ALTER TABLE neosync_api.jobs
ADD COLUMN IF NOT EXISTS run_timeout BIGINT NULL,
ADD COLUMN IF NOT EXISTS sync_options jsonb NOT NULL DEFAULT '{}'::jsonb;
