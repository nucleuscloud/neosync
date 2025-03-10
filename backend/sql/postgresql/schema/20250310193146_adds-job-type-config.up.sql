ALTER TABLE neosync_api.jobs
ADD COLUMN jobtype_config JSONB NOT NULL DEFAULT '{}'::JSONB;
