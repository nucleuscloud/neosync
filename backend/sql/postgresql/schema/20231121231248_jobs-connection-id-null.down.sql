-- this will only work if there are not-null source connection ids
ALTER TABLE neosync_api.jobs
ALTER COLUMN connection_source_id SET NOT NULL;
