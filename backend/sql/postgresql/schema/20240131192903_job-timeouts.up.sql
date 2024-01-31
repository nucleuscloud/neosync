ALTER TABLE neosync_api.jobs
ADD COLUMN IF NOT EXISTS workflow_options jsonb NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS sync_options jsonb NOT NULL DEFAULT '{}'::jsonb;
