ALTER TABLE neosync_api.jobs
ADD COLUMN IF NOT EXISTS connection_source_id uuid null;

ALTER TABLE neosync_api.jobs
DROP CONSTRAINT IF EXISTS fk_jobs_conn_source_id_conn_id;

ALTER TABLE neosync_api.jobs
ADD CONSTRAINT fk_jobs_conn_source_id_conn_id FOREIGN KEY (connection_source_id) REFERENCES neosync_api.connections(id);
