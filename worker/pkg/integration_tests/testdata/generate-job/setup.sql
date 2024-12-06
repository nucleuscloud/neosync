CREATE SCHEMA IF NOT EXISTS generate_job;

CREATE TABLE IF NOT EXISTS generate_job.regions (
	region_id SERIAL PRIMARY KEY,
	region_name CHARACTER VARYING (25)
);
