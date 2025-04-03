ALTER TABLE neosync_api.job_hooks
  DROP CONSTRAINT IF EXISTS job_hooks_name_unique;

ALTER TABLE neosync_api.job_hooks
  ADD CONSTRAINT job_hooks_name_unique UNIQUE (job_id, name);
