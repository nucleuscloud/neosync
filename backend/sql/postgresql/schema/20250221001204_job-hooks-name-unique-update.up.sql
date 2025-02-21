DROP INDEX IF EXISTS job_hooks_name_unique;

CREATE UNIQUE INDEX job_hooks_name_unique ON job_hooks (name, job_id);
