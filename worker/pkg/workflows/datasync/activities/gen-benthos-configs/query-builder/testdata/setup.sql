CREATE SCHEMA IF NOT EXISTS genbenthosconfigs_querybuilder;

SET search_path TO genbenthosconfigs_querybuilder;


 CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- company
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
-- market
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


-- COMPANY DATA
INSERT INTO company (name, url, employee_count, uuid)
VALUES
  ('Acme Corporation', 'www.acme.com', 500, uuid_generate_v4()),
  ('Global Enterprises', 'globalenterprises.net', 1200, uuid_generate_v4()),
  ('Tech Innovations', 'www.techinnovations.io', 250, uuid_generate_v4());

-- DEPARTMENT DATA 
INSERT INTO department (name, url, company_id, uuid)
VALUES
  ('Marketing', 'marketing.acme.com', 1, uuid_generate_v4()),  -- Acme Corporation
  ('Sales', 'sales.acme.com', 1, uuid_generate_v4()), 
  ('Finance', null, 2, uuid_generate_v4()), -- Global Enterprises
  ('R&D', 'rnd.techinnovations.io', 3, uuid_generate_v4()); -- Tech Innovations

-- TRANSACTION DATA
INSERT INTO transaction (id, amount, created, updated, department_id, date, currency, description, uuid)
VALUES
  (1, 250.50, now() - interval '2 weeks', now(), 1, '2024-05-01', 'USD', 'Office Supplies', uuid_generate_v4()),
  (2, 1250.00, now() - interval '5 days', now(), 2, '2024-05-06', 'GBP', 'Travel Expenses', uuid_generate_v4()),
  (3, 87.25, now() - interval '1 month', now(), 3, '2024-04-20', 'EUR', 'Lunch Meeting', uuid_generate_v4());

-- EXPENSE REPORT DATA
INSERT INTO expense_report (id, invoice_id, date, amount, department_source_id, department_destination_id, created, updated, currency, transaction_type, paid, adjustment_amount)
VALUES
  (1, 'INV-1234', '2024-05-03', 500.00, 1, 2, now() - interval '15 days', now(), 'USD', 1, true, null),
  (2,'INV-5678', '2024-04-28', 128.75, 3, 1, now() - interval '20 days', now() - interval '1 day', 'CAD', 2, false, 12.50);


 update expense_report set transaction_id = 1 where id = 1;
  update expense_report set transaction_id = 3 where id = 2;
