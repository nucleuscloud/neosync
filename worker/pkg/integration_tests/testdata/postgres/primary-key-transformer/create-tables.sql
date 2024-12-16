CREATE TABLE IF NOT EXISTS store_notifications (

	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid()
);

CREATE TABLE IF NOT EXISTS stores (

	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
notifications_id uuid UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS store_customers (

	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
store_id uuid NOT NULL,
referred_by_code uuid NULL
);

CREATE TABLE IF NOT EXISTS referral_codes (

	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
customer_id uuid NOT NULL
);

ALTER TABLE store_customers ADD FOREIGN KEY (store_id) REFERENCES stores (id);


ALTER TABLE store_customers ADD FOREIGN KEY (referred_by_code) REFERENCES referral_codes (id);


ALTER TABLE stores ADD FOREIGN KEY (notifications_id) REFERENCES store_notifications (id);


ALTER TABLE referral_codes ADD FOREIGN KEY (customer_id) REFERENCES store_customers (id);




INSERT INTO store_notifications (id) VALUES

('c0e8db5a-1b71-4e5b-8d7b-6f5a1a4d0736'), ('d37fa1b8-4a36-42a6-8786-3a5f5641c302'), ('f8d7a2b0-9d73-47b2-91f8-8e10c37c2e74'),
('a7e4b9f9-9c6c-4b48-9f98-3e1a4b7d2c1f'), ('b6c3d7a4-5e6a-4a0d-8f72-9e3b1a7c9d4f'), ('e4d7a6b8-7e5c-4a3b-9f6d-5e2a1c3b7a4d'),
('f3a7d9b2-8d6b-4a1d-9e5b-6a3f1d7c2e0a'), ('d4b8c3e1-6e7a-4b9f-8a7c-5e3d2c1b9a0f'), ('e7c3a4d5-1f8b-4a2d-9e3f-7b6a4d2c9e1f'),
('c3e8b1d6-7a5b-4f9e-8d6c-3a2b1d7c4e9a'), ('a9f8d4b7-3a2e-4f1b-9d6c-7e5a3b2c1f4e'), ('d1a7f9b3-6e5c-4a2b-9d8f-3a7c4e2b1d6a'),
('e9a4b3d6-7f5b-4e2d-9c6a-5d3b1f2a9c7e'), ('b1e8d3a4-2f7c-4a5b-9d8f-6e3a1c7b2d4e'), ('d7b8e1a3-6c5f-4b9e-8d7a-3f2c1e9b7a4d'),
('e4d8c3a1-7f5b-4a2e-9d6c-5a3f1b7c2e9d'), ('c9a8e7b3-1d6f-4a2b-9d5e-7c3a2f1e6b4d'), ('a3d8b1e4-6f5c-4a9b-8d7e-5b2c1f3a7d4e'),
('d6b7e4a9-7f5d-4a2b-9c6e-3f2a1e8d5b9c'), ('e1a8c3d4-2f6b-4a5e-9d8c-7b3d1f2a6e5d');


INSERT INTO stores (id, notifications_id) VALUES

('a1b2c3d4-e5f6-47a8-9b0c-d1e2f3a4b5c6', 'c0e8db5a-1b71-4e5b-8d7b-6f5a1a4d0736'),
('b2c3d4e5-f6a7-48b9-0c1d-e2f3a4b5c6d7', 'd37fa1b8-4a36-42a6-8786-3a5f5641c302'),
('c3d4e5f6-a7b8-49c0-1d2e-f3a4b5c6d7e8', 'f8d7a2b0-9d73-47b2-91f8-8e10c37c2e74'),
('d4e5f6a7-b8c9-40d1-2e3f-a4b5c6d7e8f9', 'a7e4b9f9-9c6c-4b48-9f98-3e1a4b7d2c1f'),
('e5f6a7b8-c9d0-41e2-3f4a-b5c6d7e8f9a1', 'b6c3d7a4-5e6a-4a0d-8f72-9e3b1a7c9d4f'),
('f6a7b8c9-d0e1-42f3-4a5b-c6d7e8f9a1b2', 'e4d7a6b8-7e5c-4a3b-9f6d-5e2a1c3b7a4d'),
('a7b8c9d0-e1f2-43a4-5b6c-d7e8f9a1b2c3', 'f3a7d9b2-8d6b-4a1d-9e5b-6a3f1d7c2e0a'),
('b8c9d0e1-f2a3-44b5-6c7d-e8f9a1b2c3d4', 'd4b8c3e1-6e7a-4b9f-8a7c-5e3d2c1b9a0f'),
('c9d0e1f2-a3b4-45c6-7d8e-f9a1b2c3d4e5', 'e7c3a4d5-1f8b-4a2d-9e3f-7b6a4d2c9e1f'),
('d0e1f2a3-b4c5-46d7-8e9f-a1b2c3d4e5f6', 'c3e8b1d6-7a5b-4f9e-8d6c-3a2b1d7c4e9a'),
('e1f2a3b4-c5d6-47e8-9f0a-b2c3d4e5f6a7', 'a9f8d4b7-3a2e-4f1b-9d6c-7e5a3b2c1f4e'),
('f2a3b4c5-d6e7-48f9-0a1b-c3d4e5f6a7b8', 'd1a7f9b3-6e5c-4a2b-9d8f-3a7c4e2b1d6a'),
('a3b4c5d6-e7f8-49a0-1b2c-d4e5f6a7b8c9', 'e9a4b3d6-7f5b-4e2d-9c6a-5d3b1f2a9c7e'),
('b4c5d6e7-f8a9-40b1-2c3d-e5f6a7b8c9d0', 'b1e8d3a4-2f7c-4a5b-9d8f-6e3a1c7b2d4e'),
('c5d6e7f8-a9b0-41c2-3d4e-f6a7b8c9d0e1', 'd7b8e1a3-6c5f-4b9e-8d7a-3f2c1e9b7a4d'),
('d6e7f8a9-b0c1-42d3-4e5f-a7b8c9d0e1f2', 'e4d8c3a1-7f5b-4a2e-9d6c-5a3f1b7c2e9d'),
('e7f8a9b0-c1d2-43e4-5f6a-b8c9d0e1f2a3', 'c9a8e7b3-1d6f-4a2b-9d5e-7c3a2f1e6b4d'),
('f8a9b0c1-d2e3-44f5-6a7b-c9d0e1f2a3b4', 'a3d8b1e4-6f5c-4a9b-8d7e-5b2c1f3a7d4e'),
('a9b0c1d2-e3f4-45a6-7b8c-d0e1f2a3b4c5', 'd6b7e4a9-7f5d-4a2b-9c6e-3f2a1e8d5b9c'),
('b0c1d2e3-f4a5-46b7-8c9d-e1f2a3b4c5d6', 'e1a8c3d4-2f6b-4a5e-9d8c-7b3d1f2a6e5d');


INSERT INTO store_customers (id, store_id, referred_by_code) VALUES

('c1d2e3f4-a5b6-47c8-9d0e-f1a2b3c4d5e6', 'a1b2c3d4-e5f6-47a8-9b0c-d1e2f3a4b5c6', NULL),
('d2e3f4a5-b6c7-48d9-0e1f-a2b3c4d5e6f7', 'b2c3d4e5-f6a7-48b9-0c1d-e2f3a4b5c6d7', NULL),
('e3f4a5b6-c7d8-49e0-1f2a-b3c4d5e6f7a8', 'c3d4e5f6-a7b8-49c0-1d2e-f3a4b5c6d7e8', NULL),
('f4a5b6c7-d8e9-40f1-2a3b-c4d5e6f7a8b9', 'd4e5f6a7-b8c9-40d1-2e3f-a4b5c6d7e8f9', NULL),
('a5b6c7d8-e9f0-41a2-3b4c-d5e6f7a8b9c0', 'e5f6a7b8-c9d0-41e2-3f4a-b5c6d7e8f9a1', NULL),
('b6c7d8e9-f0a1-42b3-4c5d-e6f7a8b9c0d1', 'f6a7b8c9-d0e1-42f3-4a5b-c6d7e8f9a1b2', NULL),
('c7d8e9f0-a1b2-43c4-5d6e-f7a8b9c0d1e2', 'a7b8c9d0-e1f2-43a4-5b6c-d7e8f9a1b2c3', NULL),
('d8e9f0a1-b2c3-44d5-6e7f-a8b9c0d1e2f3', 'b8c9d0e1-f2a3-44b5-6c7d-e8f9a1b2c3d4', NULL),
('e9f0a1b2-c3d4-45e6-7f8a-b9c0d1e2f3a4', 'c9d0e1f2-a3b4-45c6-7d8e-f9a1b2c3d4e5', NULL),
('f0a1b2c3-d4e5-46f7-8a9b-c0d1e2f3a4b5', 'd0e1f2a3-b4c5-46d7-8e9f-a1b2c3d4e5f6', NULL),
('a1b2c3d4-e5f6-47a8-9b0c-d1e2f3a4b5c6', 'e1f2a3b4-c5d6-47e8-9f0a-b2c3d4e5f6a7', NULL),
('b2c3d4e5-f6a7-48b9-0c1d-e2f3a4b5c6d7', 'f2a3b4c5-d6e7-48f9-0a1b-c3d4e5f6a7b8', NULL),
('c3d4e5f6-a7b8-49c0-1d2e-f3a4b5c6d7e8', 'a3b4c5d6-e7f8-49a0-1b2c-d4e5f6a7b8c9', NULL),
('d4e5f6a7-b8c9-40d1-2e3f-a4b5c6d7e8f9', 'b4c5d6e7-f8a9-40b1-2c3d-e5f6a7b8c9d0', NULL),
('e5f6a7b8-c9d0-41e2-3f4a-b5c6d7e8f9a1', 'c5d6e7f8-a9b0-41c2-3d4e-f6a7b8c9d0e1', NULL),
('f6a7b8c9-d0e1-42f3-4a5b-c6d7e8f9a1b2', 'd6e7f8a9-b0c1-42d3-4e5f-a7b8c9d0e1f2', NULL),
('a7b8c9d0-e1f2-43a4-5b6c-d7e8f9a1b2c3', 'e7f8a9b0-c1d2-43e4-5f6a-b8c9d0e1f2a3', NULL),
('b8c9d0e1-f2a3-44b5-6c7d-e8f9a1b2c3d4', 'f8a9b0c1-d2e3-44f5-6a7b-c9d0e1f2a3b4', NULL),
('c9d0e1f2-a3b4-45c6-7d8e-f9a1b2c3d4e5', 'a9b0c1d2-e3f4-45a6-7b8c-d0e1f2a3b4c5', NULL),
('d0e1f2a3-b4c5-46d7-8e9f-a1b2c3d4e5f6', 'b0c1d2e3-f4a5-46b7-8c9d-e1f2a3b4c5d6', NULL);


INSERT INTO referral_codes (id, customer_id) VALUES

('e2f3a4b5-c6d7-48e9-0a1b-f1a2b3c4d5e6', 'c1d2e3f4-a5b6-47c8-9d0e-f1a2b3c4d5e6'),
('f3a4b5c6-d7e8-49f0-1b2c-a2b3c4d5e6f7', 'd2e3f4a5-b6c7-48d9-0e1f-a2b3c4d5e6f7'),
('a4b5c6d7-e8f9-40a1-2b3c-b3c4d5e6f7a8', 'e3f4a5b6-c7d8-49e0-1f2a-b3c4d5e6f7a8'),
('b5c6d7e8-f9a0-41b2-3c4d-c4d5e6f7a8b9', 'f4a5b6c7-d8e9-40f1-2a3b-c4d5e6f7a8b9'),
('c6d7e8f9-a0b1-42c3-4d5e-d5e6f7a8b9c0', 'a5b6c7d8-e9f0-41a2-3b4c-d5e6f7a8b9c0'),
('d7e8f9a0-b1c2-43d4-5e6f-e6f7a8b9c0d1', 'b6c7d8e9-f0a1-42b3-4c5d-e6f7a8b9c0d1'),
('e8f9a0b1-c2d3-44e5-6f7a-f7a8b9c0d1e2', 'c7d8e9f0-a1b2-43c4-5d6e-f7a8b9c0d1e2'),
('f9a0b1c2-d3e4-45f6-7a8b-a8b9c0d1e2f3', 'd8e9f0a1-b2c3-44d5-6e7f-a8b9c0d1e2f3'),
('a0b1c2d3-e4f5-46a7-8b9c-b9c0d1e2f3a4', 'e9f0a1b2-c3d4-45e6-7f8a-b9c0d1e2f3a4'),
('b1c2d3e4-f5a6-47b8-9c0d-c0d1e2f3a4b5', 'f0a1b2c3-d4e5-46f7-8a9b-c0d1e2f3a4b5'),
('c2d3e4f5-a6b7-48c9-0d1e-d1e2f3a4b5c6', 'a1b2c3d4-e5f6-47a8-9b0c-d1e2f3a4b5c6'),
('d3e4f5a6-b7c8-49d0-1e2f-e2f3a4b5c6d7', 'b2c3d4e5-f6a7-48b9-0c1d-e2f3a4b5c6d7'),
('e4f5a6b7-c8d9-40e1-2f3a-f3a4b5c6d7e8', 'c3d4e5f6-a7b8-49c0-1d2e-f3a4b5c6d7e8'),
('f5a6b7c8-d9e0-41f2-3a4b-a4b5c6d7e8f9', 'd4e5f6a7-b8c9-40d1-2e3f-a4b5c6d7e8f9'),
('a6b7c8d9-e0f1-42a3-4b5c-b5c6d7e8f9a0', 'e5f6a7b8-c9d0-41e2-3f4a-b5c6d7e8f9a1'),
('b7c8d9e0-f1a2-43b4-5c6d-c6d7e8f9a1b2', 'f6a7b8c9-d0e1-42f3-4a5b-c6d7e8f9a1b2'),
('c8d9e0f1-a2b3-44c5-6d7e-d7e8f9a1b2c3', 'a7b8c9d0-e1f2-43a4-5b6c-d7e8f9a1b2c3'),
('d9e0f1a2-b3c4-45d6-7e8f-e8f9a1b2c3d4', 'b8c9d0e1-f2a3-44b5-6c7d-e8f9a1b2c3d4'),
('e0f1a2b3-c4d5-46e7-8f9a-f9a1b2c3d4e5', 'c9d0e1f2-a3b4-45c6-7d8e-f9a1b2c3d4e5'),
('f1a2b3c4-d5e6-47f8-9a0b-a1b2c3d4e5f6', 'd0e1f2a3-b4c5-46d7-8e9f-a1b2c3d4e5f6');


UPDATE store_customers SET referred_by_code = 'e2f3a4b5-c6d7-48e9-0a1b-f1a2b3c4d5e6' WHERE id = 'c1d2e3f4-a5b6-47c8-9d0e-f1a2b3c4d5e6';

UPDATE store_customers SET referred_by_code = 'f3a4b5c6-d7e8-49f0-1b2c-a2b3c4d5e6f7' WHERE id = 'd2e3f4a5-b6c7-48d9-0e1f-a2b3c4d5e6f7';

UPDATE store_customers SET referred_by_code = 'a4b5c6d7-e8f9-40a1-2b3c-b3c4d5e6f7a8' WHERE id = 'e3f4a5b6-c7d8-49e0-1f2a-b3c4d5e6f7a8';

UPDATE store_customers SET referred_by_code = 'b5c6d7e8-f9a0-41b2-3c4d-c4d5e6f7a8b9' WHERE id = 'f4a5b6c7-d8e9-40f1-2a3b-c4d5e6f7a8b9';

UPDATE store_customers SET referred_by_code = 'c6d7e8f9-a0b1-42c3-4d5e-d5e6f7a8b9c0' WHERE id = 'a5b6c7d8-e9f0-41a2-3b4c-d5e6f7a8b9c0';

UPDATE store_customers SET referred_by_code = 'd7e8f9a0-b1c2-43d4-5e6f-e6f7a8b9c0d1' WHERE id = 'b6c7d8e9-f0a1-42b3-4c5d-e6f7a8b9c0d1';

UPDATE store_customers SET referred_by_code = 'e8f9a0b1-c2d3-44e5-6f7a-f7a8b9c0d1e2' WHERE id = 'c7d8e9f0-a1b2-43c4-5d6e-f7a8b9c0d1e2';

UPDATE store_customers SET referred_by_code = 'f9a0b1c2-d3e4-45f6-7a8b-a8b9c0d1e2f3' WHERE id = 'd8e9f0a1-b2c3-44d5-6e7f-a8b9c0d1e2f3';

UPDATE store_customers SET referred_by_code = 'a0b1c2d3-e4f5-46a7-8b9c-b9c0d1e2f3a4' WHERE id = 'e9f0a1b2-c3d4-45e6-7f8a-b9c0d1e2f3a4';

UPDATE store_customers SET referred_by_code = 'b1c2d3e4-f5a6-47b8-9c0d-c0d1e2f3a4b5' WHERE id = 'f0a1b2c3-d4e5-46f7-8a9b-c0d1e2f3a4b5';

UPDATE store_customers SET referred_by_code = 'c2d3e4f5-a6b7-48c9-0d1e-d1e2f3a4b5c6' WHERE id = 'a1b2c3d4-e5f6-47a8-9b0c-d1e2f3a4b5c6';

UPDATE store_customers SET referred_by_code = 'd3e4f5a6-b7c8-49d0-1e2f-e2f3a4b5c6d7' WHERE id = 'b2c3d4e5-f6a7-48b9-0c1d-e2f3a4b5c6d7';

UPDATE store_customers SET referred_by_code = 'e4f5a6b7-c8d9-40e1-2f3a-f3a4b5c6d7e8' WHERE id = 'c3d4e5f6-a7b8-49c0-1d2e-f3a4b5c6d7e8';

UPDATE store_customers SET referred_by_code = 'f5a6b7c8-d9e0-41f2-3a4b-a4b5c6d7e8f9' WHERE id = 'd4e5f6a7-b8c9-40d1-2e3f-a4b5c6d7e8f9';

UPDATE store_customers SET referred_by_code = 'a6b7c8d9-e0f1-42a3-4b5c-b5c6d7e8f9a0' WHERE id = 'e5f6a7b8-c9d0-41e2-3f4a-b5c6d7e8f9a1';

UPDATE store_customers SET referred_by_code = 'b7c8d9e0-f1a2-43b4-5c6d-c6d7e8f9a1b2' WHERE id = 'f6a7b8c9-d0e1-42f3-4a5b-c6d7e8f9a1b2';

UPDATE store_customers SET referred_by_code = 'c8d9e0f1-a2b3-44c5-6d7e-d7e8f9a1b2c3' WHERE id = 'a7b8c9d0-e1f2-43a4-5b6c-d7e8f9a1b2c3';

UPDATE store_customers SET referred_by_code = 'd9e0f1a2-b3c4-45d6-7e8f-e8f9a1b2c3d4' WHERE id = 'b8c9d0e1-f2a3-44b5-6c7d-e8f9a1b2c3d4';

UPDATE store_customers SET referred_by_code = 'e0f1a2b3-c4d5-46e7-8f9a-f9a1b2c3d4e5' WHERE id = 'c9d0e1f2-a3b4-45c6-7d8e-f9a1b2c3d4e5';

UPDATE store_customers SET referred_by_code = 'f1a2b3c4-d5e6-47f8-9a0b-a1b2c3d4e5f6' WHERE id = 'd0e1f2a3-b4c5-46d7-8e9f-a1b2c3d4e5f6';

