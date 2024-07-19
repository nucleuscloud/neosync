CREATE SCHEMA IF NOT EXISTS double_reference;
SET search_path TO double_reference;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE
	IF NOT EXISTS "company" (
		"id" BIGSERIAL NOT NULL,
		"name" text NOT NULL,
		"url" text NULL,
		"employee_count" integer NULL,
		"uuid" uuid NOT NULL DEFAULT uuid_generate_v4 (),
		CONSTRAINT company_pkey PRIMARY KEY (id),
		CONSTRAINT company_uuid_key UNIQUE (uuid)
	);
CREATE TABLE
	IF NOT EXISTS "department" (
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
