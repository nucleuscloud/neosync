ALTER TABLE neosync_api.jobs
ADD COLUMN schema_mappings jsonb,
ADD COLUMN schema_changes jsonb;
