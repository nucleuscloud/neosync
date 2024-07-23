CREATE SCHEMA IF NOT EXISTS primary_key;
SET search_path TO primary_key;

CREATE TABLE IF NOT EXISTS primary_key.store_notifications (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid()
);

CREATE TABLE IF NOT EXISTS primary_key.stores (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
notifications_id uuid UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS primary_key.store_customers (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
store_id uuid NOT NULL,
referred_by_code uuid NULL
);

CREATE TABLE IF NOT EXISTS primary_key.referral_codes (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
customer_id uuid NOT NULL
);

ALTER TABLE primary_key.store_customers ADD FOREIGN KEY (store_id) REFERENCES primary_key.stores (id);
ALTER TABLE primary_key.store_customers ADD FOREIGN KEY (referred_by_code) REFERENCES primary_key.referral_codes (id);
ALTER TABLE primary_key.stores ADD FOREIGN KEY (notifications_id) REFERENCES primary_key.store_notifications (id);
ALTER TABLE primary_key.referral_codes ADD FOREIGN KEY (customer_id) REFERENCES primary_key.store_customers (id);



