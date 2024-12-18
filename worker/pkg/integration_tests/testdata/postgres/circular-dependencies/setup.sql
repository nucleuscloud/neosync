CREATE SCHEMA IF NOT EXISTS circular_dependencies;
SET search_path TO circular_dependencies;

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
