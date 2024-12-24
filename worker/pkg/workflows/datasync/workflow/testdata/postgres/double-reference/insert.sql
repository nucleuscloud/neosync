SET search_path TO double_reference;
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
  -- Repeat with varied data ...

-- EXPENSE REPORT DATA
INSERT INTO expense_report (id, invoice_id, date, amount, department_source_id, department_destination_id, created, updated, currency, transaction_type, paid, adjustment_amount, transaction_id)
VALUES
  (1, 'INV-1234', '2024-05-03', 500.00, 1, 2, now() - interval '15 days', now(), 'USD', 1, true, null, 1),
  (2,'INV-5678', '2024-04-28', 128.75, 3, 1, now() - interval '20 days', now() - interval '1 day', 'CAD', 2, false, 12.50, 3);
