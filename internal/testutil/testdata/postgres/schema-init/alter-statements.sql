-- add columns
ALTER TABLE regions
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN region_number BIGINT GENERATED ALWAYS AS IDENTITY (START WITH 100 INCREMENT BY 10);

ALTER TABLE countries
    ADD COLUMN last_update DATE DEFAULT CURRENT_DATE;

ALTER TABLE locations
    ADD COLUMN phone_numbers TEXT[];

ALTER TABLE jobs
    ADD COLUMN job_type CHARACTER VARYING(50),
    ADD COLUMN job_code SERIAL;

ALTER TABLE employees
    ADD COLUMN profile JSONB;

CREATE SEQUENCE employees_code_seq
    START 1000
    INCREMENT 1;

ALTER TABLE employees
    ADD COLUMN employee_code INTEGER DEFAULT nextval('employees_code_seq');

COMMENT ON COLUMN employees.profile IS 'A JSONB column containing employee profile information';
COMMENT ON COLUMN countries.last_update IS 'The last time the country was updated';
COMMENT ON COLUMN jobs.job_type IS 'I''m an astronaut';

UPDATE regions
SET is_active = TRUE
WHERE region_id IN (1, 2, 3);

UPDATE regions
SET is_active = FALSE
WHERE region_id = 4;

UPDATE countries
SET last_update = '2025-01-01'
WHERE country_id IN ('US','CA');

UPDATE countries
SET last_update = '2025-02-01'
WHERE country_id IN ('UK','DE','FR');

UPDATE locations
SET phone_numbers = ARRAY['+1-555-123-4567','+1-555-999-8888']
WHERE location_id = 1400;

UPDATE locations
SET phone_numbers = ARRAY['+44-20-1111-2222']
WHERE location_id = 2400;

UPDATE jobs
SET job_type = 'Accounting'
WHERE job_id IN (1,2,6);

UPDATE jobs
SET job_type = 'Administration'
WHERE job_id IN (3,5,8);

UPDATE jobs
SET job_type = 'Management'
WHERE job_id IN (4,7,10,14,15,19);

UPDATE jobs
SET job_type = 'Technical'
WHERE job_id IN (9);

UPDATE jobs
SET job_type = 'Sales/Other'
WHERE job_id IN (11,12,13,16,17,18);

UPDATE employees
SET profile = '{"hobbies":["cycling","reading"],"location":"Remote"}'
WHERE employee_id = 100;

UPDATE employees
SET profile = '{"hobbies":["hiking"],"location":"On-site"}'
WHERE employee_id = 101;

-- Example: set a more complex JSON structure for a different employee
UPDATE employees
SET profile = '{"languages":["English","German"],"certifications":["CPA","MBA"]}'
WHERE employee_id = 108;
