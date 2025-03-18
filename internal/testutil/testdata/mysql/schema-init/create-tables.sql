CREATE TABLE IF NOT EXISTS regions (
    region_id INT (11) AUTO_INCREMENT PRIMARY KEY,
    region_name VARCHAR (25) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS countries (
    country_id CHAR (2) PRIMARY KEY,
    country_name VARCHAR (40) DEFAULT NULL,
    region_id INT (11) NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
    location_id INT (11) AUTO_INCREMENT PRIMARY KEY,
    street_address VARCHAR (40) DEFAULT NULL,
    postal_code VARCHAR (12) DEFAULT NULL,
    city VARCHAR (30) NOT NULL,
    state_province VARCHAR (25) DEFAULT NULL,
    country_id CHAR (2) NOT NULL
);

CREATE TABLE IF NOT EXISTS jobs (
    job_id INT (11) AUTO_INCREMENT PRIMARY KEY,
    job_title VARCHAR (35) NOT NULL,
    min_salary DECIMAL (8, 2) DEFAULT NULL,
    max_salary DECIMAL (8, 2) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS departments (
    department_id INT (11) AUTO_INCREMENT PRIMARY KEY,
    department_name VARCHAR (30) NOT NULL,
    location_id INT (11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS employees (
    employee_id INT (11) AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR (20) DEFAULT NULL,
    last_name VARCHAR (25) NOT NULL,
    email VARCHAR (100) NOT NULL,
    phone_number VARCHAR (20) DEFAULT NULL,
    hire_date DATE NOT NULL,
    job_id INT (11) NOT NULL,
    salary DECIMAL (8, 2) NOT NULL,
    manager_id INT (11) DEFAULT NULL,
    department_id INT (11) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS dependents (
    dependent_id INT (11) AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR (50) NOT NULL,
    last_name VARCHAR (50) NOT NULL,
    relationship VARCHAR (25) NOT NULL,
    employee_id INT (11) NOT NULL
);

CREATE TABLE IF NOT EXISTS emails (
    email VARCHAR (100) NOT NULL
);

-- Add foreign keys
ALTER TABLE countries
    ADD FOREIGN KEY (region_id) REFERENCES regions (region_id)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE locations
    ADD FOREIGN KEY (country_id) REFERENCES countries (country_id)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE departments
    ADD FOREIGN KEY (location_id) REFERENCES locations (location_id)
    ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE employees
    ADD FOREIGN KEY (job_id) REFERENCES jobs (job_id)
    ON DELETE CASCADE ON UPDATE CASCADE,
    ADD FOREIGN KEY (department_id) REFERENCES departments (department_id)
    ON DELETE CASCADE ON UPDATE CASCADE,
    ADD FOREIGN KEY (manager_id) REFERENCES employees (employee_id);

ALTER TABLE dependents
    ADD FOREIGN KEY (employee_id) REFERENCES employees (employee_id)
    ON DELETE CASCADE ON UPDATE CASCADE;

INSERT INTO emails (email) VALUES
('john.doe@example.com'),
('jane.smith@example.com'),
('alice@example.com'),
('bob@example.com'),
('charlie@example.com'),
('david@example.com'),
('eve@example.com'),
('frank@example.com'),
('grace@example.com'),
('heidi@example.com');


/*Data for the table regions */

INSERT INTO regions(region_id,region_name) VALUES (1,'Europe');
INSERT INTO regions(region_id,region_name) VALUES (2,'Americas');
INSERT INTO regions(region_id,region_name) VALUES (3,'Asia');
INSERT INTO regions(region_id,region_name) VALUES (4,'Middle East and Africa');


/*Data for the table countries */
INSERT INTO countries(country_id,country_name,region_id) VALUES ('AR','Argentina',2);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('AU','Australia',3);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('BE','Belgium',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('BR','Brazil',2);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('CA','Canada',2);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('CH','Switzerland',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('CN','China',3);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('DE','Germany',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('DK','Denmark',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('EG','Egypt',4);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('FR','France',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('HK','HongKong',3);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('IL','Israel',4);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('IN','India',3);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('IT','Italy',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('JP','Japan',3);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('KW','Kuwait',4);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('MX','Mexico',2);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('NG','Nigeria',4);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('NL','Netherlands',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('SG','Singapore',3);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('UK','United Kingdom',1);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('US','United States of America',2);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('ZM','Zambia',4);
INSERT INTO countries(country_id,country_name,region_id) VALUES ('ZW','Zimbabwe',4);

/*Data for the table locations */
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (1400,'2014 Jabberwocky Rd','26192','Southlake','Texas','US');
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (1500,'2011 Interiors Blvd','99236','South San Francisco','California','US');
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (1700,'2004 Charade Rd','98199','Seattle','Washington','US');
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (1800,'147 Spadina Ave','M5V 2L7','Toronto','Ontario','CA');
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (2400,'8204 Arthur St',NULL,'London',NULL,'UK');
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (2500,'Magdalen Centre, The Oxford Science Park','OX9 9ZB','Oxford','Oxford','UK');
INSERT INTO locations(location_id,street_address,postal_code,city,state_province,country_id) VALUES (2700,'Schwanthalerstr. 7031','80925','Munich','Bavaria','DE');


/*Data for the table jobs */

INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (1,'Public Accountant',4200.00,9000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (2,'Accounting Manager',8200.00,16000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (3,'Administration Assistant',3000.00,6000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (4,'President',20000.00,40000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (5,'Administration Vice President',15000.00,30000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (6,'Accountant',4200.00,9000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (7,'Finance Manager',8200.00,16000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (8,'Human Resources Representative',4000.00,9000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (9,'Programmer',4000.00,10000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (10,'Marketing Manager',9000.00,15000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (11,'Marketing Representative',4000.00,9000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (12,'Public Relations Representative',4500.00,10500.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (13,'Purchasing Clerk',2500.00,5500.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (14,'Purchasing Manager',8000.00,15000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (15,'Sales Manager',10000.00,20000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (16,'Sales Representative',6000.00,12000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (17,'Shipping Clerk',2500.00,5500.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (18,'Stock Clerk',2000.00,5000.00);
INSERT INTO jobs(job_id,job_title,min_salary,max_salary) VALUES (19,'Stock Manager',5500.00,8500.00);


/*Data for the table departments */

INSERT INTO departments(department_id,department_name,location_id) VALUES (1,'Administration',1700);
INSERT INTO departments(department_id,department_name,location_id) VALUES (2,'Marketing',1800);
INSERT INTO departments(department_id,department_name,location_id) VALUES (3,'Purchasing',1700);
INSERT INTO departments(department_id,department_name,location_id) VALUES (4,'Human Resources',2400);
INSERT INTO departments(department_id,department_name,location_id) VALUES (5,'Shipping',1500);
INSERT INTO departments(department_id,department_name,location_id) VALUES (6,'IT',1400);
INSERT INTO departments(department_id,department_name,location_id) VALUES (7,'Public Relations',2700);
INSERT INTO departments(department_id,department_name,location_id) VALUES (8,'Sales',2500);
INSERT INTO departments(department_id,department_name,location_id) VALUES (9,'Executive',1700);
INSERT INTO departments(department_id,department_name,location_id) VALUES (10,'Finance',1700);
INSERT INTO departments(department_id,department_name,location_id) VALUES (11,'Accounting',1700);



/*Data for the table employees */

INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (100,'Steven','King','steven.king@sqltutorial.org','515.123.4567','1987-06-17',4,24000.00,NULL,9);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (101,'Neena','Kochhar','neena.kochhar@sqltutorial.org','515.123.4568','1989-09-21',5,17000.00,100,9);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (102,'Lex','De Haan','lex.de haan@sqltutorial.org','515.123.4569','1993-01-13',5,17000.00,100,9);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (103,'Alexander','Hunold','alexander.hunold@sqltutorial.org','590.423.4567','1990-01-03',9,9000.00,102,6);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (104,'Bruce','Ernst','bruce.ernst@sqltutorial.org','590.423.4568','1991-05-21',9,6000.00,103,6);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (105,'David','Austin','david.austin@sqltutorial.org','590.423.4569','1997-06-25',9,4800.00,103,6);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (106,'Valli','Pataballa','valli.pataballa@sqltutorial.org','590.423.4560','1998-02-05',9,4800.00,103,6);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (107,'Diana','Lorentz','diana.lorentz@sqltutorial.org','590.423.5567','1999-02-07',9,4200.00,103,6);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (108,'Nancy','Greenberg','nancy.greenberg@sqltutorial.org','515.124.4569','1994-08-17',7,12000.00,101,10);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (109,'Daniel','Faviet','daniel.faviet@sqltutorial.org','515.124.4169','1994-08-16',6,9000.00,108,10);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (110,'John','Chen','john.chen@sqltutorial.org','515.124.4269','1997-09-28',6,8200.00,108,10);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (111,'Ismael','Sciarra','ismael.sciarra@sqltutorial.org','515.124.4369','1997-09-30',6,7700.00,108,10);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (112,'Jose Manuel','Urman','jose manuel.urman@sqltutorial.org','515.124.4469','1998-03-07',6,7800.00,108,10);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (113,'Luis','Popp','luis.popp@sqltutorial.org','515.124.4567','1999-12-07',6,6900.00,108,10);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (114,'Den','Raphaely','den.raphaely@sqltutorial.org','515.127.4561','1994-12-07',14,11000.00,100,3);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (115,'Alexander','Khoo','alexander.khoo@sqltutorial.org','515.127.4562','1995-05-18',13,3100.00,114,3);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (116,'Shelli','Baida','shelli.baida@sqltutorial.org','515.127.4563','1997-12-24',13,2900.00,114,3);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (117,'Sigal','Tobias','sigal.tobias@sqltutorial.org','515.127.4564','1997-07-24',13,2800.00,114,3);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (118,'Guy','Himuro','guy.himuro@sqltutorial.org','515.127.4565','1998-11-15',13,2600.00,114,3);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (119,'Karen','Colmenares','karen.colmenares@sqltutorial.org','515.127.4566','1999-08-10',13,2500.00,114,3);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (120,'Matthew','Weiss','matthew.weiss@sqltutorial.org','650.123.1234','1996-07-18',19,8000.00,100,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (121,'Adam','Fripp','adam.fripp@sqltutorial.org','650.123.2234','1997-04-10',19,8200.00,100,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (122,'Payam','Kaufling','payam.kaufling@sqltutorial.org','650.123.3234','1995-05-01',19,7900.00,100,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (123,'Shanta','Vollman','shanta.vollman@sqltutorial.org','650.123.4234','1997-10-10',19,6500.00,100,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (126,'Irene','Mikkilineni','irene.mikkilineni@sqltutorial.org','650.124.1224','1998-09-28',18,2700.00,120,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (145,'John','Russell','john.russell@sqltutorial.org',NULL,'1996-10-01',15,14000.00,100,8);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (146,'Karen','Partners','karen.partners@sqltutorial.org',NULL,'1997-01-05',15,13500.00,100,8);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (176,'Jonathon','Taylor','jonathon.taylor@sqltutorial.org',NULL,'1998-03-24',16,8600.00,100,8);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (177,'Jack','Livingston','jack.livingston@sqltutorial.org',NULL,'1998-04-23',16,8400.00,100,8);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (178,'Kimberely','Grant','kimberely.grant@sqltutorial.org',NULL,'1999-05-24',16,7000.00,100,8);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (179,'Charles','Johnson','charles.johnson@sqltutorial.org',NULL,'2000-01-04',16,6200.00,100,8);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (192,'Sarah','Bell','sarah.bell@sqltutorial.org','650.501.1876','1996-02-04',17,4000.00,123,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (193,'Britney','Everett','britney.everett@sqltutorial.org','650.501.2876','1997-03-03',17,3900.00,123,5);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (200,'Jennifer','Whalen','jennifer.whalen@sqltutorial.org','515.123.4444','1987-09-17',3,4400.00,101,1);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (201,'Michael','Hartstein','michael.hartstein@sqltutorial.org','515.123.5555','1996-02-17',10,13000.00,100,2);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (202,'Pat','Fay','pat.fay@sqltutorial.org','603.123.6666','1997-08-17',11,6000.00,201,2);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (203,'Susan','Mavris','susan.mavris@sqltutorial.org','515.123.7777','1994-06-07',8,6500.00,101,4);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (204,'Hermann','Baer','hermann.baer@sqltutorial.org','515.123.8888','1994-06-07',12,10000.00,101,7);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (205,'Shelley','Higgins','shelley.higgins@sqltutorial.org','515.123.8080','1994-06-07',2,12000.00,101,11);
INSERT INTO employees(employee_id,first_name,last_name,email,phone_number,hire_date,job_id,salary,manager_id,department_id) VALUES (206,'William','Gietz','william.gietz@sqltutorial.org','515.123.8181','1994-06-07',1,8300.00,205,11);


/*Data for the table dependents */

INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (1,'Penelope','Gietz','Child',206);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (2,'Nick','Higgins','Child',205);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (3,'Ed','Whalen','Child',200);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (4,'Jennifer','King','Child',100);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (5,'Johnny','Kochhar','Child',101);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (6,'Bette','De Haan','Child',102);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (7,'Grace','Faviet','Child',109);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (8,'Matthew','Chen','Child',110);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (9,'Joe','Sciarra','Child',111);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (10,'Christian','Urman','Child',112);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (11,'Zero','Popp','Child',113);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (12,'Karl','Greenberg','Child',108);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (13,'Uma','Mavris','Child',203);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (14,'Vivien','Hunold','Child',103);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (15,'Cuba','Ernst','Child',104);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (16,'Fred','Austin','Child',105);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (17,'Helen','Pataballa','Child',106);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (18,'Dan','Lorentz','Child',107);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (19,'Bob','Hartstein','Child',201);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (20,'Lucille','Fay','Child',202);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (21,'Kirsten','Baer','Child',204);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (22,'Elvis','Khoo','Child',115);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (23,'Sandra','Baida','Child',116);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (24,'Cameron','Tobias','Child',117);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (25,'Kevin','Himuro','Child',118);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (26,'Rip','Colmenares','Child',119);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (27,'Julia','Raphaely','Child',114);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (28,'Woody','Russell','Child',145);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (29,'Alec','Partners','Child',146);
INSERT INTO dependents(dependent_id,first_name,last_name,relationship,employee_id) VALUES (30,'Sandra','Taylor','Child',176);



-- ==================================================
-- 1) Create a 'grandparent' table
-- ==================================================
CREATE TABLE IF NOT EXISTS grandparent (
  gp_id INT AUTO_INCREMENT,
  gp_name VARCHAR(50) NOT NULL,
  PRIMARY KEY (gp_id)
);

-- Insert a few sample rows
INSERT INTO grandparent (gp_name) VALUES
('Grandparent A'),
('Grandparent B'),
('Grandparent C');

-- ==================================================
-- 2) Create a 'parent' table referencing 'grandparent'
-- ==================================================
CREATE TABLE IF NOT EXISTS parent (
  p_id INT AUTO_INCREMENT,
  p_unique_val VARCHAR(50) NOT NULL,
  gp_id INT NOT NULL,  -- references grandparent
  PRIMARY KEY (p_id),
  UNIQUE KEY ux_parent_unique (p_unique_val),
  CONSTRAINT fk_parent_gp
    FOREIGN KEY (gp_id)
    REFERENCES grandparent (gp_id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);

-- Insert sample data
INSERT INTO parent (p_unique_val, gp_id) VALUES
('Parent A - unique', 1),
('Parent B - unique', 2),
('Parent C - unique', 3);

-- ==================================================
-- 3) Create a 'child' table referencing BOTH 'parent' AND 'grandparent'
--    This ensures there's a chain: grandparent -> parent -> child,
--    plus a direct grandparent -> child.
-- ==================================================
CREATE TABLE IF NOT EXISTS child (
  c_id INT AUTO_INCREMENT,
  c_data VARCHAR(100),
  p_id INT NOT NULL,  -- references parent
  gp_id INT NOT NULL, -- also references grandparent
  PRIMARY KEY (c_id),
  CONSTRAINT fk_child_parent
    FOREIGN KEY (p_id)
    REFERENCES parent (p_id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
  CONSTRAINT fk_child_gp
    FOREIGN KEY (gp_id)
    REFERENCES grandparent (gp_id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);

-- Sample data
INSERT INTO child (c_data, p_id, gp_id) VALUES
('Child of Parent A / Grandparent A', 1, 1),
('Child of Parent B / Grandparent B', 2, 2),
('Child of Parent C / Grandparent C', 3, 3);

-- ==================================================
-- 4) OPTIONAL: Create a 'cyclic_table' with a self-referencing FK
--    This ensures you have a scenario of dropping a self-join FK.
-- ==================================================
CREATE TABLE IF NOT EXISTS cyclic_table (
  cycle_id INT AUTO_INCREMENT,
  cycle_name VARCHAR(50) NOT NULL,
  manager_cycle_id INT DEFAULT NULL,
  PRIMARY KEY (cycle_id),
  CONSTRAINT fk_cycle_self
    FOREIGN KEY (manager_cycle_id)
    REFERENCES cyclic_table (cycle_id)
    ON DELETE SET NULL
    ON UPDATE CASCADE
);

-- Insert sample data
INSERT INTO cyclic_table (cycle_name, manager_cycle_id) VALUES
('Top Node', NULL),     -- top-level
('Mid Node', 1),        -- managed by 'Top Node'
('Leaf Node', 2);       -- managed by 'Mid Node'

-- ==================================================
-- 5) 'multi_col_parent' with a composite PK
-- ==================================================
CREATE TABLE IF NOT EXISTS multi_col_parent (
  mcp_a INT NOT NULL,
  mcp_b INT NOT NULL,
  description VARCHAR(50),
  PRIMARY KEY (mcp_a, mcp_b)  -- composite PK
);

-- Insert sample data
INSERT INTO multi_col_parent (mcp_a, mcp_b, description) VALUES
(1, 100, 'Multi-Col Parent #1'),
(2, 200, 'Multi-Col Parent #2');

-- ==================================================
-- 6) 'multi_col_child' referencing multi_col_parent AND parent
--    Also add a CHECK constraint for MySQL 8.0.16+.
-- ==================================================
CREATE TABLE IF NOT EXISTS multi_col_child (
  mc_child_id INT AUTO_INCREMENT,
  mcp_a INT NOT NULL,
  mcp_b INT NOT NULL,
  p_id_ref INT NOT NULL,  -- also references parent
  some_value INT NOT NULL DEFAULT 10,
  PRIMARY KEY (mc_child_id),

  -- Foreign Key referencing the composite PK of multi_col_parent
  CONSTRAINT fk_mcc_mcp
    FOREIGN KEY (mcp_a, mcp_b)
    REFERENCES multi_col_parent (mcp_a, mcp_b)
    ON DELETE RESTRICT
    ON UPDATE CASCADE,

  -- Another foreign key referencing 'parent'
  CONSTRAINT fk_mcc_parent
    FOREIGN KEY (p_id_ref)
    REFERENCES parent (p_id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,

  -- Check constraint (requires 8.0.16+ to enforce)
  CONSTRAINT chk_some_value
    CHECK (some_value <= 50)
);

-- Insert sample data
INSERT INTO multi_col_child (mcp_a, mcp_b, p_id_ref, some_value)
VALUES
 (1, 100, 1, 20),
 (2, 200, 2, 40);

-- trigger test
CREATE TABLE IF NOT EXISTS astronaut (
    astronaut_id INT AUTO_INCREMENT PRIMARY KEY,
    full_name   VARCHAR(100),
    position    VARCHAR(50)
);


CREATE TABLE IF NOT EXISTS astronaut_log (
    log_id      INT AUTO_INCREMENT PRIMARY KEY,
    astronaut_id INT,
    full_name   VARCHAR(100),
    action      VARCHAR(20),
    logged_at   DATETIME
);


CREATE TRIGGER astronaut_ai
	AFTER UPDATE ON astronaut
	FOR EACH ROW
BEGIN
	UPDATE
		astronaut_log
	SET
		full_name = NEW.full_name,
		action = 'UPDATED',
		logged_at = NOW()
	WHERE
		astronaut_id = NEW.astronaut_id;
END;

INSERT INTO astronaut (full_name, position)
VALUES
    ('Neil Armstrong', 'Astronaut'),
    ('Buzz Aldrin', 'Astronaut');

INSERT INTO astronaut_log (
    astronaut_id,
    full_name,
    action,
    logged_at
)
VALUES
    (1, 'Neil Armstrong', 'INSERTED', NOW()),
    (2, 'Buzz Aldrin', 'INSERTED', NOW());
