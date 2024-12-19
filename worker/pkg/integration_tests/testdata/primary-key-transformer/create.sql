CREATE SCHEMA IF NOT EXISTS "primary_$key";
SET search_path TO "primary_$key";

CREATE TABLE IF NOT EXISTS "primary_$key".store_notifications (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid()
);

CREATE TABLE IF NOT EXISTS "primary_$key".stores (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
notifications_id uuid UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "primary_$key".store_customers (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
store_id uuid NOT NULL,
referred_by_code uuid NULL
);

CREATE TABLE IF NOT EXISTS "primary_$key".referral_codes (
	id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
customer_id uuid NOT NULL
);

ALTER TABLE "primary_$key".store_customers ADD FOREIGN KEY (store_id) REFERENCES "primary_$key".stores (id);
ALTER TABLE "primary_$key".store_customers ADD FOREIGN KEY (referred_by_code) REFERENCES "primary_$key".referral_codes (id);
ALTER TABLE "primary_$key".stores ADD FOREIGN KEY (notifications_id) REFERENCES "primary_$key".store_notifications (id);
ALTER TABLE "primary_$key".referral_codes ADD FOREIGN KEY (customer_id) REFERENCES "primary_$key".store_customers (id);



