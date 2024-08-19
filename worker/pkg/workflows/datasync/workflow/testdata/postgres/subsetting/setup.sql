CREATE SCHEMA IF NOT EXISTS subsetting;

SET search_path TO subsetting;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";



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

