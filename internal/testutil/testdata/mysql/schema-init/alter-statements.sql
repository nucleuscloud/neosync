-- ============================================
-- 1) Add new columns to existing tables
-- ============================================
ALTER TABLE regions
  ADD COLUMN region_code CHAR(3) NOT NULL DEFAULT 'UNK';

ALTER TABLE countries
  ADD COLUMN updated_at DATETIME NOT NULL
    DEFAULT CURRENT_TIMESTAMP
    ON UPDATE CURRENT_TIMESTAMP;

ALTER TABLE locations
  ADD COLUMN established_year YEAR NOT NULL DEFAULT 2000;

ALTER TABLE departments
  ADD COLUMN dept_label VARCHAR(50)
    GENERATED ALWAYS AS ( CONCAT('DEPT-', department_name) ) VIRTUAL;

-- Add an auto-increment "identity" to emails
-- (Requires MySQL 8.0+ for an AUTO_INCREMENT primary key on an existing table)
ALTER TABLE emails
  ADD COLUMN email_identity INT NOT NULL AUTO_INCREMENT PRIMARY KEY;
ALTER TABLE emails AUTO_INCREMENT = 100;

ALTER TABLE employees
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1,
  ADD COLUMN last_modified TIMESTAMP NOT NULL
    DEFAULT CURRENT_TIMESTAMP
    ON UPDATE CURRENT_TIMESTAMP,
  ADD COLUMN full_name VARCHAR(46)
    GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) STORED,
  ADD COLUMN extra_info JSON DEFAULT NULL;  -- (MySQL 5.7+)

ALTER TABLE dependents
  ADD COLUMN added_on DATE NOT NULL DEFAULT '2020-01-01';


ALTER TABLE employees
  DROP FOREIGN KEY employees_ibfk_3;


-- Now optionally drop the manager_id column if you want:
ALTER TABLE employees
  DROP COLUMN manager_id;

-- ============================================
-- 3) Add a brand-new foreign key
--    (For example, a new column employees.country_id -> countries.country_id)
-- ============================================
ALTER TABLE employees
  ADD COLUMN country_id CHAR(2);

ALTER TABLE employees
  ADD FOREIGN KEY (country_id) REFERENCES countries(country_id)
    ON DELETE SET NULL
    ON UPDATE CASCADE;

-- ============================================
-- 4) Add a CHECK constraint (MySQL 8.0.16+)
--    Enforce that salary must always be > 0
-- ============================================
ALTER TABLE employees
  ADD CONSTRAINT check_salary
  CHECK (salary > 0);

-- ============================================
-- 5) Drop a column from locations
-- ============================================
ALTER TABLE locations
  DROP COLUMN state_province;


-- ============================================
-- 7) Example updates to test new columns
-- ============================================
UPDATE regions
  SET region_code = 'EUR'
  WHERE region_name = 'Europe';

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

-- ================================================
-- 1) Dropping constraints from multi_col_child
--    in correct order before dropping PK from multi_col_parent
-- ================================================
ALTER TABLE multi_col_child
  DROP FOREIGN KEY fk_mcc_parent;    -- references parent.p_id
ALTER TABLE multi_col_child
  DROP FOREIGN KEY fk_mcc_mcp;       -- references multi_col_parent(mcp_a, mcp_b)
ALTER TABLE multi_col_child
  DROP CHECK chk_some_value;


-- ================================================
-- 2) Now we can safely drop the composite PK in multi_col_parent
--    Because no table references it now
-- ================================================
ALTER TABLE multi_col_parent
  DROP PRIMARY KEY;


-- ================================================
-- 3) Dropping the self-reference in 'cyclic_table'
-- ================================================
ALTER TABLE cyclic_table
  DROP FOREIGN KEY fk_cycle_self;


-- ================================================
-- 4) Dropping constraints in 'child'
--    'child' references 'parent' and 'grandparent'
-- ================================================
ALTER TABLE child
  DROP FOREIGN KEY fk_child_parent;
ALTER TABLE child
  DROP FOREIGN KEY fk_child_gp;

-- ================================================
-- 5) Dropping constraints in 'parent'
--    'parent' references 'grandparent'
--    Also has a unique key on p_unique_val
-- ================================================
ALTER TABLE parent
  DROP FOREIGN KEY fk_parent_gp;

-- Drop the UNIQUE constraint by dropping the index
ALTER TABLE parent
  DROP INDEX ux_parent_unique;

DROP TRIGGER IF EXISTS astronaut_ai;

CREATE TRIGGER astronaut_ai
	AFTER UPDATE ON astronaut
	FOR EACH ROW
BEGIN
	UPDATE
		astronaut_log
	SET
		full_name = CONCAT(NEW.full_name, ' (', NEW.position, ')'),  -- Updated logic
		action = 'UPDATED',
		logged_at = NOW()
	WHERE
		astronaut_id = NEW.astronaut_id;
END;
