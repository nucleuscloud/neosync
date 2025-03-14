-- 1) regions: Add region_code
ALTER TABLE regions
  ADD COLUMN region_code CHAR(3) NOT NULL DEFAULT 'UNK';

-- 2) countries: Add updated_at
ALTER TABLE countries
  ADD COLUMN updated_at DATETIME NOT NULL 
    DEFAULT CURRENT_TIMESTAMP 
    ON UPDATE CURRENT_TIMESTAMP;

-- 3) locations: Add established_year
ALTER TABLE locations
  ADD COLUMN established_year YEAR NOT NULL DEFAULT 2000;

-- 4) departments: Add a virtual generated column
ALTER TABLE departments
  ADD COLUMN dept_label VARCHAR(50) 
    GENERATED ALWAYS AS ( CONCAT('DEPT-', department_name) ) VIRTUAL;

-- 5) emails: Add an identity column (MySQL 8.0+ only)
ALTER TABLE emails
  ADD COLUMN email_identity INT NOT NULL AUTO_INCREMENT PRIMARY KEY;
ALTER TABLE emails AUTO_INCREMENT = 100;


-- 6) employees: Add multiple columns

ALTER TABLE employees
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD COLUMN last_modified TIMESTAMP NOT NULL 
      DEFAULT CURRENT_TIMESTAMP 
      ON UPDATE CURRENT_TIMESTAMP,
  ADD COLUMN full_name VARCHAR(46) 
      GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) STORED;


-- 7) employees: Add a JSON column (MySQL 5.7+)
ALTER TABLE employees
  ADD COLUMN extra_info JSON DEFAULT NULL;

-- 8) dependents: Add default date
ALTER TABLE dependents
  ADD COLUMN added_on DATE NOT NULL DEFAULT '2020-01-01';


-- Example updates for region_code (CHAR(3))
UPDATE regions
  SET region_code = 'EUR'
  WHERE region_name = 'Europe';

UPDATE regions
  SET region_code = 'AME'
  WHERE region_name = 'Americas';

UPDATE regions
  SET region_code = 'ASI'
  WHERE region_name = 'Asia';

UPDATE regions
  SET region_code = 'MEA'
  WHERE region_name = 'Middle East and Africa';



-- Force a specific updated_at for these countries
UPDATE countries
  SET updated_at = '2022-12-31 10:00:00'
  WHERE country_id IN ('AR','AU','BE');


UPDATE locations
  SET established_year = 1990
  WHERE location_id = 1400;

UPDATE locations
  SET established_year = 2005
  WHERE location_id = 1500;

UPDATE locations
  SET established_year = 1985
  WHERE location_id = 1700;

UPDATE locations
  SET established_year = 1970
  WHERE location_id = 1800;

UPDATE locations
  SET established_year = 2000
  WHERE location_id = 2400;

UPDATE locations
  SET established_year = 2021
  WHERE location_id = 2500;

UPDATE locations
  SET established_year = 1965
  WHERE location_id = 2700;


-- Mark one employee as inactive
UPDATE employees
  SET is_active = 0
  WHERE employee_id = 100;

-- Mark a few employees specifically as active
UPDATE employees
  SET is_active = 1
  WHERE employee_id IN (101, 102, 103);


UPDATE employees
  SET last_modified = '2023-01-01 12:00:00'
  WHERE employee_id = 100;


UPDATE employees
  SET extra_info = JSON_OBJECT('hobbies', 'reading', 'favorite_color', 'blue')
  WHERE employee_id = 108;

UPDATE employees
  SET extra_info = JSON_OBJECT('skills', JSON_ARRAY('SQL','Java','Python'))
  WHERE employee_id = 103;


-- Assign 'added_on' dates to some dependents
UPDATE dependents
  SET added_on = '2023-01-15'
  WHERE dependent_id = 1;

UPDATE dependents
  SET added_on = '2023-01-20'
  WHERE dependent_id = 2;

UPDATE dependents
  SET added_on = '2023-02-01'
  WHERE dependent_id = 3;

