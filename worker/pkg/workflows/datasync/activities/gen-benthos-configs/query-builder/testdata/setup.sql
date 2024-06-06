CREATE SCHEMA IF NOT EXISTS genbenthosconfigs_querybuilder;

SET search_path TO genbenthosconfigs_querybuilder;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Test_BuildQueryMap_DoubleReference
CREATE TABLE
	IF NOT EXISTS company (
		"id" BIGSERIAL NOT NULL,
		"name" text NOT NULL,
		"url" text NULL,
		"employee_count" integer NULL,
		"uuid" uuid NOT NULL DEFAULT uuid_generate_v4 (),
		CONSTRAINT company_pkey PRIMARY KEY (id),
		CONSTRAINT company_uuid_key UNIQUE (uuid)
	);
CREATE TABLE
	IF NOT EXISTS department (
		"id" BIGSERIAL NOT NULL,
		"name" text NOT NULL,
		"url" text NULL,
		"company_id" bigint NOT NULL, -- to be fk
		"user_id" bigint NULL,
		"uuid" uuid NOT NULL DEFAULT uuid_generate_v4 (),
		CONSTRAINT department_pkey PRIMARY KEY (id),
		CONSTRAINT department_uuid_key UNIQUE (uuid),
		CONSTRAINT department_company_id_fkey FOREIGN KEY (company_id) REFERENCES company (id) ON DELETE CASCADE
	);
CREATE TABLE IF NOT EXISTS transaction (
    id bigint NOT NULL,
    amount double precision NOT NULL,
    created timestamp without time zone,
    updated timestamp without time zone,
  	department_id bigint, -- to be fk
    date date,
    currency text NOT NULL,
    settings json DEFAULT '{
      "historicalCount": 0,
      "vacation": false,
      "conference": true,
      "travel": true
    }'::json NOT NULL,
    description text,
    timezone text DEFAULT 'America/New_York'::text NOT NULL,
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
  	CONSTRAINT transaction_pkey PRIMARY KEY (id),
  	CONSTRAINT transaction_department_id_fkey FOREIGN KEY (department_id) REFERENCES department (ID) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS expense_report (
    id bigint NOT NULL,
    invoice_id text,
    date date NOT NULL,
    amount numeric(15,2),
    department_source_id bigint, -- fk
    department_destination_id bigint, --fk
    created timestamp without time zone,
    updated timestamp without time zone,
    currency character varying(5),
    transaction_type integer NOT NULL,
    paid boolean DEFAULT false,
    transaction_id bigint, -- fk
    adjustment_amount numeric(15,2),
    CONSTRAINT transaction_type_valid_values CHECK ((transaction_type = ANY (ARRAY[1, 2]))),
  	CONSTRAINT expense_report_pkey PRIMARY KEY (id),
  	CONSTRAINT expense_report_dept_source_fkey FOREIGN KEY (department_source_id) REFERENCES department (ID) ON DELETE CASCADE,
  	CONSTRAINT expense_report_dept_destination_fkey FOREIGN KEY (department_destination_id) REFERENCES department (ID) ON DELETE CASCADE,
  	CONSTRAINT expense_report_transaction_fkey FOREIGN KEY (transaction_id) REFERENCES transaction (ID) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS expense (
    id bigint NOT NULL PRIMARY KEY,
    report_id bigint,
  	CONSTRAINT expense_report_d_fkey FOREIGN KEY (report_id) REFERENCES expense_report (ID) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS item (
    id bigint NOT NULL PRIMARY KEY,
    expense_id bigint,
  	CONSTRAINT expense_id_fkey FOREIGN KEY (expense_id) REFERENCES expense (ID) ON DELETE CASCADE
);


INSERT INTO company (name, url, employee_count, uuid)
VALUES
  ('Acme Corporation', 'www.acme.com', 500, uuid_generate_v4()),
  ('Global Enterprises', 'globalenterprises.net', 1200, uuid_generate_v4()),
  ('Tech Innovations', 'www.techinnovations.io', 250, uuid_generate_v4());

INSERT INTO department (name, url, company_id, uuid)
VALUES
  ('Marketing', 'marketing.acme.com', 1, uuid_generate_v4()),  -- Acme Corporation
  ('Sales', 'sales.acme.com', 1, uuid_generate_v4()), 
  ('Finance', null, 2, uuid_generate_v4()), -- Global Enterprises
  ('R&D', 'rnd.techinnovations.io', 3, uuid_generate_v4()); -- Tech Innovations

INSERT INTO transaction (id, amount, created, updated, department_id, date, currency, description, uuid)
VALUES
  (1, 250.50, now() - interval '2 weeks', now(), 1, '2024-05-01', 'USD', 'Office Supplies', uuid_generate_v4()),
  (2, 1250.00, now() - interval '5 days', now(), 2, '2024-05-06', 'GBP', 'Travel Expenses', uuid_generate_v4()),
  (3, 87.25, now() - interval '1 month', now(), 3, '2024-04-20', 'EUR', 'Lunch Meeting', uuid_generate_v4());

INSERT INTO expense_report (id, invoice_id, date, amount, department_source_id, department_destination_id, created, updated, currency, transaction_type, paid, adjustment_amount, transaction_id)
VALUES
  (1, 'INV-1234', '2024-05-03', 500.00, 1, 2, now() - interval '15 days', now(), 'USD', 1, true, null, 1),
  (2,'INV-5678', '2024-04-28', 128.75, 3, 1, now() - interval '20 days', now() - interval '1 day', 'CAD', 2, false, 12.50, 3),
  (3,'INV-5678', '2024-04-28', 128.75, 2, 1, now() - interval '20 days', now() - interval '1 day', 'CAD', 2, false, 12.50, 2);


INSERT INTO expense (id, report_id) VALUES 
(1, 2), 
(2, 3), 
(3, 1);

-- Insert statements for item
INSERT INTO item (id, expense_id) VALUES 
(1, 3), 
(2, 1), 
(3, 2);


-- Test_BuildQueryMap_DoubleRootSubset
CREATE TABLE test_2_x (
  id BIGINT NOT NULL PRIMARY KEY,
  name text,
  created timestamp without time zone
);

CREATE TABLE test_2_b (
  id BIGINT NOT NULL PRIMARY KEY,
  name text,
  created timestamp without time zone
);

CREATE TABLE test_2_a (
  id BIGINT NOT NULL PRIMARY KEY,
  x_id BIGINT NOT NULL,
  CONSTRAINT test2_x FOREIGN KEY (x_id) REFERENCES test_2_x (id) ON DELETE CASCADE
);

CREATE TABLE test_2_c (
  id BIGINT NOT NULL PRIMARY KEY,
  name text,
  created timestamp without time zone,
  a_id BIGINT NOT NULL,
  b_id BIGINT NOT NULL,
  CONSTRAINT test2_a FOREIGN KEY (a_id) REFERENCES test_2_a (id) ON DELETE CASCADE,
  CONSTRAINT test2_b FOREIGN KEY (b_id) REFERENCES test_2_b (id) ON DELETE CASCADE
);

CREATE TABLE test_2_d (
  id BIGINT NOT NULL PRIMARY KEY,
  c_id BIGINT NOT NULL,
  CONSTRAINT test2_x FOREIGN KEY (c_id) REFERENCES test_2_c (id) ON DELETE CASCADE
);


CREATE TABLE test_2_e (
  id BIGINT NOT NULL PRIMARY KEY,
  c_id BIGINT NOT NULL,
  CONSTRAINT test2_x FOREIGN KEY (c_id) REFERENCES test_2_c (id) ON DELETE CASCADE
);

INSERT INTO test_2_x (id, name, created) VALUES 
(1, 'Xander', '2023-06-01 10:00:00'),
(2, 'Xena', '2023-06-02 11:00:00'),
(3, 'Xavier', '2023-06-03 12:00:00'),
(4, 'Xiomara', '2023-06-04 13:00:00'),
(5, 'Xaviera', '2023-06-05 14:00:00');

INSERT INTO test_2_b (id, name, created) VALUES 
(1, 'Beta1', '2023-06-01 15:00:00'),
(2, 'Beta2', '2023-06-02 16:00:00'),
(3, 'Beta3', '2023-06-03 17:00:00'),
(4, 'Beta4', '2023-06-04 18:00:00'),
(5, 'Beta5', '2023-06-05 19:00:00');

INSERT INTO test_2_a (id, x_id) VALUES 
(1, 5),
(2, 4),
(3, 3),
(4, 3),
(5, 1);

INSERT INTO test_2_c (id, name, created, a_id, b_id) VALUES 
(1, 'Gamma1', '2023-06-01 20:00:00', 1, 1),
(2, 'Gamma2', '2023-06-02 21:00:00', 2, 2),
(3, 'Gamma3', '2023-06-03 22:00:00', 3, 3),
(4, 'Gamma4', '2023-06-04 23:00:00', 4, 4),
(5, 'Gamma5', '2023-06-05 00:00:00', 5, 5);

INSERT INTO test_2_d (id, c_id) VALUES 
(1, 1),
(2, 2),
(3, 3),
(4, 4),
(5, 5);

INSERT INTO test_2_e (id, c_id) VALUES 
(1, 1),
(2, 2),
(3, 3),
(4, 4),
(5, 5);

-- Test_BuildQueryMap_MultipleRoots, Test_BuildQueryMap_MultipleSubset, Test_BuildQueryMap_MultipleSubsets_SubsetsByForeignKeysOff
CREATE TABLE test_3_a (
  id BIGINT NOT NULL PRIMARY KEY
);
CREATE TABLE test_3_b (
  id BIGINT NOT NULL PRIMARY KEY,
  a_id BIGINT NOT NULL,
  CONSTRAINT test3_a FOREIGN KEY (a_id) REFERENCES test_3_a (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_c (
  id BIGINT NOT NULL PRIMARY KEY,
  b_id BIGINT NOT NULL,
  CONSTRAINT test3_b FOREIGN KEY (b_id) REFERENCES test_3_b (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_d (
  id BIGINT NOT NULL PRIMARY KEY,
  c_id BIGINT NOT NULL,
  CONSTRAINT test3_c FOREIGN KEY (c_id) REFERENCES test_3_c (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_e (
  id BIGINT NOT NULL PRIMARY KEY,
  d_id BIGINT NOT NULL,
  CONSTRAINT test3_d FOREIGN KEY (d_id) REFERENCES test_3_d (id) ON DELETE CASCADE
);
CREATE TABLE test_3_f (
  id BIGINT NOT NULL PRIMARY KEY
);
CREATE TABLE test_3_g (
  id BIGINT NOT NULL PRIMARY KEY,
  f_id BIGINT NOT NULL,
  CONSTRAINT test3_f FOREIGN KEY (f_id) REFERENCES test_3_f (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_h (
  id BIGINT NOT NULL PRIMARY KEY,
  g_id BIGINT NOT NULL,
  CONSTRAINT test3_g FOREIGN KEY (g_id) REFERENCES test_3_g (id) ON DELETE CASCADE
);
CREATE TABLE test_3_i (
  id BIGINT NOT NULL PRIMARY KEY,
  h_id BIGINT NOT NULL,
  CONSTRAINT test3_h FOREIGN KEY (h_id) REFERENCES test_3_h (id) ON DELETE CASCADE
);

INSERT INTO test_3_a (id) VALUES 
(1), 
(2), 
(3), 
(4), 
(5);
INSERT INTO test_3_b (id, a_id) VALUES 
(1, 3), 
(2, 5), 
(3, 1), 
(4, 4), 
(5, 2);
INSERT INTO test_3_c (id, b_id) VALUES 
(1, 2), 
(2, 4), 
(3, 1), 
(4, 3), 
(5, 5);
INSERT INTO test_3_d (id, c_id) VALUES 
(1, 5), 
(2, 1), 
(3, 4), 
(4, 2), 
(5, 3);
INSERT INTO test_3_e (id, d_id) VALUES 
(1, 2), 
(2, 4), 
(3, 1), 
(4, 5), 
(5, 3);
INSERT INTO test_3_f (id) VALUES 
(1), 
(2), 
(3), 
(4), 
(5);
INSERT INTO test_3_g (id, f_id) VALUES 
(1, 5), 
(2, 1), 
(3, 4), 
(4, 2), 
(5, 3);
INSERT INTO test_3_h (id, g_id) VALUES 
(1, 4), 
(2, 2), 
(3, 5), 
(4, 1), 
(5, 3);
INSERT INTO test_3_i (id,h_id) VALUES 
(1, 4), 
(8, 2), 
(3, 3), 
(9, 1), 
(5, 3);

-- circular dependency tests
CREATE TABLE addresses (
    id BIGINT NOT NULL PRIMARY KEY,
    order_id BIGINT NULL  
);

CREATE TABLE customers (
    id BIGINT NOT NULL PRIMARY KEY,
    address_id BIGINT,
    CONSTRAINT fk_address
        FOREIGN KEY (address_id) 
        REFERENCES addresses (id)
);

CREATE TABLE orders (
  id BIGINT NOT NULL PRIMARY KEY,
    customer_id BIGINT,
    CONSTRAINT fk_customer
        FOREIGN KEY (customer_id) 
        REFERENCES customers (id)
);

CREATE TABLE payments (
  id BIGINT NOT NULL PRIMARY KEY,
    customer_id BIGINT,
    CONSTRAINT fk_customer
        FOREIGN KEY (customer_id) 
        REFERENCES customers (id)
);

INSERT INTO addresses (id, order_id) VALUES 
(1, 1), 
(2, 2), 
(3, 3), 
(4, 4), 
(5, 5);

INSERT INTO customers (id, address_id) VALUES 
(1, 3), 
(2, 5), 
(3, 1), 
(4, 4), 
(5, 2);

INSERT INTO orders (id, customer_id) VALUES 
(1, 5), 
(2, 3), 
(3, 1), 
(4, 4), 
(5, 2);

INSERT INTO payments (id, customer_id) VALUES 
(1, 4), 
(2, 2), 
(3, 1);


-- Adding the foreign key constraint to create the circular dependency
ALTER TABLE addresses
ADD CONSTRAINT fk_order
FOREIGN KEY (order_id) 
REFERENCES orders (id);


-- composite keys
CREATE TABLE division (
    id BIGINT PRIMARY KEY,
    division_name VARCHAR(100),
    location VARCHAR(100)
);
CREATE TABLE employees (
    id BIGINT,
    division_id BIGINT,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(100),
    PRIMARY KEY (id, division_id),
    FOREIGN KEY (division_id) REFERENCES division (id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);
CREATE TABLE projects (
    id BIGINT PRIMARY KEY,
    project_name VARCHAR(100),
    start_date DATE,
    end_date DATE,
    responsible_employee_id BIGINT,
    responsible_division_id BIGINT,
    CONSTRAINT FK_Projects_Employees FOREIGN KEY (responsible_employee_id, responsible_division_id)
    REFERENCES employees (id, division_id)
    ON DELETE SET NULL
    ON UPDATE CASCADE
);

INSERT INTO division (id, division_name, location) VALUES
(1, 'Marketing', 'New York'),
(2, 'Finance', 'London'),
(3, 'Human Resources', 'San Francisco'),
(4, 'IT', 'Berlin'),
(5, 'Customer Service', 'Tokyo');


INSERT INTO employees (id, division_id, first_name, last_name, email) VALUES
(6, 1, 'Alice', 'Johnson', 'alice.johnson@example.com'),
(7, 2, 'Bob', 'Smith', 'bob.smith@example.com'),
(8, 3, 'Carol', 'Martinez', 'carol.martinez@example.com'),
(9, 4, 'David', 'Lee', 'david.lee@example.com'),
(10, 5, 'Eva', 'Kim', 'eva.kim@example.com');


INSERT INTO projects (id, project_name, start_date, end_date, responsible_employee_id, responsible_division_id) VALUES
(11, 'Website Redesign', '2023-05-01', '2023-10-01', 6, 1),
(12, 'Financial Audit', '2023-06-15', '2023-07-15', 7, 2),
(13, 'Hiring Initiative', '2023-09-01', '2024-01-31', 8, 3),
(14, 'Software Development', '2023-05-20', '2023-12-20', 9, 4),
(15, 'Customer Feedback Analysis', '2023-07-01', '2023-11-30', 10, 5);


-- self referencing 
CREATE TABLE bosses (
	id BIGINT PRIMARY KEY,
	manager_id BIGINT,
	big_boss_id BIGINT,
	FOREIGN KEY (manager_id) REFERENCES bosses (id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (big_boss_id) REFERENCES bosses (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE minions (
	id BIGINT PRIMARY KEY,
	boss_id BIGINT,
	FOREIGN KEY (boss_id) REFERENCES bosses (id) ON UPDATE CASCADE ON DELETE CASCADE
);

INSERT INTO bosses (id, manager_id, big_boss_id) VALUES 
(1, NULL, NULL), 
(2, 1, NULL), 
(3, 2, 1), 
(4, 3, 2), 
(5, 4, 3),
(6,NULL,NULL);

INSERT INTO minions (id, boss_id) VALUES 
(1, 4), 
(2, 3), 
(3, 1), 
(4, 5), 
(5, 2);
