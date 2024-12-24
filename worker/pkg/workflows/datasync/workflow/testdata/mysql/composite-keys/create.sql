CREATE DATABASE IF NOT EXISTS composite;
USE composite;

CREATE TABLE IF NOT EXISTS orders (
    order_id CHAR(36)  NOT NULL,
    product_id CHAR(36)  NOT NULL,
    order_date DATE,
    PRIMARY KEY (order_id, product_id)
);


CREATE TABLE IF NOT EXISTS order_details (
    detail_id CHAR(36) NOT NULL,
    order_id CHAR(36) NOT NULL,
    product_id CHAR(36) NOT NULL,
    quantity INT,
    PRIMARY KEY (detail_id)
);


CREATE TABLE IF NOT EXISTS order_shipping (
    shipping_id CHAR(36) NOT NULL,
    order_id CHAR(36) NOT NULL,
    product_id CHAR(36) NOT NULL,
    shipping_date DATE,
    PRIMARY KEY (shipping_id)
);

CREATE TABLE IF NOT EXISTS shipping_status (
    status_id CHAR(36) NOT NULL,
    order_id CHAR(36) NOT NULL,
    product_id CHAR(36) NOT NULL,
     status VARCHAR(255),
    status_date DATE,
    PRIMARY KEY (status_id)
);



ALTER TABLE order_details
ADD FOREIGN KEY (order_id, product_id)
REFERENCES orders (order_id, product_id);


ALTER TABLE order_shipping
ADD FOREIGN KEY (order_id, product_id)
REFERENCES orders (order_id, product_id);


ALTER TABLE shipping_status
ADD FOREIGN KEY (order_id, product_id)
REFERENCES order_shipping (order_id, product_id);

