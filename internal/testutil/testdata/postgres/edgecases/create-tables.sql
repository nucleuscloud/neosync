
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA public;
CREATE TABLE IF NOT EXISTS "BadName" (
    "ID" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "NAME" text UNIQUE
);

INSERT INTO "BadName" ("NAME")
VALUES 
    ('Xk7pQ9nM3v'),
    ('Rt5wLjH2yB'),
    ('Zc8fAe4dN6'),
    ('Ym9gKu3sW5'),
    ('Vb4nPx7tJ2');

CREATE TABLE "Bad Name 123!@#" (
    "ID" SERIAL PRIMARY KEY,
    "NAME" text REFERENCES "BadName" ("NAME")
);


INSERT INTO "Bad Name 123!@#" ("NAME")
VALUES 
    ('Xk7pQ9nM3v'),
    ('Rt5wLjH2yB'),
    ('Zc8fAe4dN6'),
    ('Ym9gKu3sW5'),
    ('Vb4nPx7tJ2');


-- Table addresses depends on Table orders
CREATE TABLE addresses(
    id UUID PRIMARY KEY,
    order_id UUID NULL  
);

-- Table customers depends on Table addresses
CREATE TABLE customers(
    id UUID PRIMARY KEY,
    address_id UUID,
    CONSTRAINT fk_address
        FOREIGN KEY (address_id) 
        REFERENCES addresses (id)

);

-- Table orders depends on Table customers
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id UUID,
    CONSTRAINT fk_customer
        FOREIGN KEY (customer_id) 
        REFERENCES customers (id)
);

-- Inserts for the addresses table
INSERT INTO addresses (id, order_id) VALUES ('ec3cbe4f-217e-49e0-bc7d-d0c334cc7d3b', 'f216a6f8-3bcd-46d8-8b99-e3b31dd5e6f3');
INSERT INTO addresses (id, order_id) VALUES ('f6c7d5b6-9140-4dcb-bc34-648fcb2c8d1f', 'd5e95c33-08c5-4098-8b69-4b1a8e4c6a96');
INSERT INTO addresses (id, order_id) VALUES ('cfba342e-46e5-45d7-8437-7fd7c22dfe0c', '7eaa7688-3730-4a55-8741-d6ae58a1b843');
INSERT INTO addresses (id, order_id) VALUES ('a29487f8-ec77-4f84-bb1d-4f526457baba', '47873236-95b4-4c0f-ae45-7b19d9de4abf');
INSERT INTO addresses (id, order_id) VALUES ('5c0f798e-8d4a-4f26-9b5d-1181f1b4d7a5', '9a2a85d2-1554-420b-b4e2-c769c674dbb1');
INSERT INTO addresses (id, order_id) VALUES ('e295f80d-2f60-41a0-9945-7dbff521b193', 'b63f0e2c-f6d2-472f-b41c-9d5e64b48c3c');
INSERT INTO addresses (id, order_id) VALUES ('f1a3c6e8-dccf-46c8-a0f3-79b93f3d2b0b', '6a1c1a7e-3e5c-4828-8228-91ff0b8d03e3');
INSERT INTO addresses (id, order_id) VALUES ('36f594af-6d53-4a48-a9b7-b889e2df349e', 'ec5f8a5f-7352-4e4c-9d3f-08e4dbf98df5');


-- Inserts for the customers table
INSERT INTO customers (id, address_id) VALUES ('e1a65af8-b0c2-42a0-99c4-7a91a0b2a80d', 'ec3cbe4f-217e-49e0-bc7d-d0c334cc7d3b');
INSERT INTO customers (id, address_id) VALUES ('a0e78f88-6b48-4d97-8a8a-212bece329b7', 'f6c7d5b6-9140-4dcb-bc34-648fcb2c8d1f');
INSERT INTO customers (id, address_id) VALUES ('b5c6f69e-13da-4f60-9b8d-dae2fa526b1f', 'cfba342e-46e5-45d7-8437-7fd7c22dfe0c');
INSERT INTO customers (id, address_id) VALUES ('d82c4a97-00ef-4e0b-8d3a-1b6a4e587524', 'a29487f8-ec77-4f84-bb1d-4f526457baba');
INSERT INTO customers (id, address_id) VALUES ('4b60d61b-fd6b-4d8e-8978-63c2edb9a274', '5c0f798e-8d4a-4f26-9b5d-1181f1b4d7a5');
INSERT INTO customers (id, address_id) VALUES ('76b2f70b-ade3-4d57-8b3b-fd1ccf9a5c3a', 'e295f80d-2f60-41a0-9945-7dbff521b193');
INSERT INTO customers (id, address_id) VALUES ('b83f99cc-8655-4639-9b0f-0d0c60f3a8c3', 'f1a3c6e8-dccf-46c8-a0f3-79b93f3d2b0b');
INSERT INTO customers (id, address_id) VALUES ('cf769742-74fa-4f2f-8580-df47cc927ba1', '36f594af-6d53-4a48-a9b7-b889e2df349e');
INSERT INTO customers (id, address_id) VALUES ('dd1b75e6-062d-4fbb-963d-973b612a20c1', 'a29487f8-ec77-4f84-bb1d-4f526457baba');
INSERT INTO customers (id, address_id) VALUES ('6f4587e5-4dfd-4e8d-8d98-c9c0e1b9ef2e', 'e295f80d-2f60-41a0-9945-7dbff521b193');

-- Inserts for the orders table
INSERT INTO orders (id, customer_id) VALUES ('f216a6f8-3bcd-46d8-8b99-e3b31dd5e6f3', 'e1a65af8-b0c2-42a0-99c4-7a91a0b2a80d');
INSERT INTO orders (id, customer_id) VALUES ('d5e95c33-08c5-4098-8b69-4b1a8e4c6a96', 'a0e78f88-6b48-4d97-8a8a-212bece329b7');
INSERT INTO orders (id, customer_id) VALUES ('7eaa7688-3730-4a55-8741-d6ae58a1b843', 'b5c6f69e-13da-4f60-9b8d-dae2fa526b1f');
INSERT INTO orders (id, customer_id) VALUES ('47873236-95b4-4c0f-ae45-7b19d9de4abf', 'd82c4a97-00ef-4e0b-8d3a-1b6a4e587524');
INSERT INTO orders (id, customer_id) VALUES ('9a2a85d2-1554-420b-b4e2-c769c674dbb1', '4b60d61b-fd6b-4d8e-8978-63c2edb9a274');
INSERT INTO orders (id, customer_id) VALUES ('b63f0e2c-f6d2-472f-b41c-9d5e64b48c3c', '76b2f70b-ade3-4d57-8b3b-fd1ccf9a5c3a');
INSERT INTO orders (id, customer_id) VALUES ('6a1c1a7e-3e5c-4828-8228-91ff0b8d03e3', 'b83f99cc-8655-4639-9b0f-0d0c60f3a8c3');
INSERT INTO orders (id, customer_id) VALUES ('ec5f8a5f-7352-4e4c-9d3f-08e4dbf98df5', 'cf769742-74fa-4f2f-8580-df47cc927ba1');
INSERT INTO orders (id, customer_id) VALUES ('58dca9d5-8500-4f8b-a3f3-75b6390e3c1a', 'dd1b75e6-062d-4fbb-963d-973b612a20c1');
INSERT INTO orders (id, customer_id) VALUES ('762b3bb2-3723-4e3b-8b53-b2e8057896ab', '6f4587e5-4dfd-4e8d-8d98-c9c0e1b9ef2e');



-- Adding the foreign key constraints to create the circular dependency
ALTER TABLE addresses
ADD CONSTRAINT fk_order
FOREIGN KEY (order_id) 
REFERENCES orders (id);


CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE
	IF NOT EXISTS "company" (
		"id" BIGSERIAL NOT NULL,
		"name" text NOT NULL,
		"url" text NULL,
		"employee_count" integer NULL,
		"uuid" uuid NOT NULL DEFAULT public.uuid_generate_v4 (),
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
		"uuid" uuid NOT NULL DEFAULT public.uuid_generate_v4 (),
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
    uuid uuid DEFAULT public.uuid_generate_v4() NOT NULL,
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
  ('Acme Corporation', 'www.acme.com', 500, public.uuid_generate_v4()),
  ('Global Enterprises', 'globalenterprises.net', 1200, public.uuid_generate_v4()),
  ('Tech Innovations', 'www.techinnovations.io', 250, public.uuid_generate_v4());

-- DEPARTMENT DATA 
INSERT INTO department (name, url, company_id, uuid)
VALUES
  ('Marketing', 'marketing.acme.com', 1, public.uuid_generate_v4()),  -- Acme Corporation
  ('Sales', 'sales.acme.com', 1, public.uuid_generate_v4()), 
  ('Finance', null, 2, public.uuid_generate_v4()), -- Global Enterprises
  ('R&D', 'rnd.techinnovations.io', 3, public.uuid_generate_v4()); -- Tech Innovations

-- TRANSACTION DATA
INSERT INTO transaction (id, amount, created, updated, department_id, date, currency, description, uuid)
VALUES
  (1, 250.50, now() - interval '2 weeks', now(), 1, '2024-05-01', 'USD', 'Office Supplies', public.uuid_generate_v4()),
  (2, 1250.00, now() - interval '5 days', now(), 2, '2024-05-06', 'GBP', 'Travel Expenses', public.uuid_generate_v4()),
  (3, 87.25, now() - interval '1 month', now(), 3, '2024-04-20', 'EUR', 'Lunch Meeting', public.uuid_generate_v4());
  -- Repeat with varied data ...

-- EXPENSE REPORT DATA
INSERT INTO expense_report (id, invoice_id, date, amount, department_source_id, department_destination_id, created, updated, currency, transaction_type, paid, adjustment_amount, transaction_id)
VALUES
  (1, 'INV-1234', '2024-05-03', 500.00, 1, 2, now() - interval '15 days', now(), 'USD', 1, true, null, 1),
  (2,'INV-5678', '2024-04-28', 128.75, 3, 1, now() - interval '20 days', now() - interval '1 day', 'CAD', 2, false, 12.50, 3);
