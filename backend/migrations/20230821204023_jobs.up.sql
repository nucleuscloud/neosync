CREATE TABLE IF NOT EXISTS neosync_api.jobs (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  name varchar NOT NULL,
  account_id uuid NOT NULL,

  status smallint NOT NULL,

  connection_source_id uuid NOT NULL,
  connection_options jsonb NOT NULL DEFAULT '{}'::jsonb,

  mappings jsonb NOT NULL DEFAULT '[]'::jsonb,

  cron_schedule varchar NULL,

  created_by_id uuid NOT NULL,
  updated_by_id uuid NOT NULL,

  CONSTRAINT jobs_pkey PRIMARY KEY (id),
  CONSTRAINT fk_jobs_accounts_id FOREIGN KEY (account_id) REFERENCES neosync_api.accounts(id) ON DELETE CASCADE,
  CONSTRAINT jobs_name_account_id UNIQUE (name, account_id),
  CONSTRAINT fk_jobs_created_by_users_id FOREIGN KEY (created_by_id) REFERENCES neosync_api.users(id),
  CONSTRAINT fk_jobs_updated_by_users_id FOREIGN KEY (updated_by_id) REFERENCES neosync_api.users(id),
  CONSTRAINT fk_jobs_conn_source_id_conn_id FOREIGN KEY (connection_source_id) REFERENCES neosync_api.connections(id)
);
ALTER TABLE neosync_api.jobs OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.jobs TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.jobs TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.jobs TO neosync_api_readonly;

CREATE TABLE IF NOT EXISTS neosync_api.job_destination_connection_associations (
  id uuid NOT NULL DEFAULT gen_random_uuid(),
	created_at timestamp NOT NULL DEFAULT now(),
	updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

  job_id uuid NOT NULL,
  connection_id uuid NOT NULL,
  options jsonb NOT NULL DEFAULT '{}'::jsonb,

  CONSTRAINT job_destination_connection_associations_pkey PRIMARY KEY (id),
  CONSTRAINT fk_jobdstconassoc_job_id_jobs_id FOREIGN KEY (job_id) REFERENCES neosync_api.jobs(id) ON DELETE CASCADE,
  CONSTRAINT fk_jobdstconassoc_conn_id_conn_id FOREIGN KEY (connection_id) REFERENCES neosync_api.connections(id) ON DELETE CASCADE,
  CONSTRAINT job_id_connection_id UNIQUE (job_id, connection_id)
);
ALTER TABLE neosync_api.job_destination_connection_associations OWNER TO neosync_api_owner;
GRANT ALL ON TABLE neosync_api.job_destination_connection_associations TO neosync_api_owner;
GRANT INSERT, DELETE, UPDATE, SELECT ON TABLE neosync_api.job_destination_connection_associations TO neosync_api_readwrite;
GRANT SELECT ON TABLE neosync_api.job_destination_connection_associations TO neosync_api_readonly;
