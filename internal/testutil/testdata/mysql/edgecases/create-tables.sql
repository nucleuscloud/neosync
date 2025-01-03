

CREATE TABLE IF NOT EXISTS container_status (
	id int NOT NULL AUTO_INCREMENT,
	PRIMARY KEY (id)) ENGINE = InnoDB AUTO_INCREMENT = 2 DEFAULT CHARSET = utf8mb3;

CREATE TABLE IF NOT EXISTS container (
	id int NOT NULL AUTO_INCREMENT,
	code varchar(32) NOT NULL,
	container_status_id int NOT NULL,
PRIMARY KEY (id),
UNIQUE KEY container_code_uniq (code),
KEY container_container_status_fk (container_status_id));

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    fullname varchar(101) GENERATED ALWAYS AS (CONCAT(first_name,' ',last_name)),
    age int NOT NULL,
    current_salary float NOT NULL
);

CREATE TABLE IF NOT EXISTS unique_emails (
     id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) not null UNIQUE
);

CREATE TABLE IF NOT EXISTS unique_emails_and_usernames (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    CONSTRAINT unique_email_username UNIQUE (email, username)
);


-- Testing basic circular dependencies
CREATE TABLE IF NOT EXISTS t1 (
    a INT AUTO_INCREMENT PRIMARY KEY,
    b INT NULL
);

CREATE TABLE IF NOT EXISTS t2 (
    a INT AUTO_INCREMENT PRIMARY KEY,
    b INT NULL
);

CREATE TABLE IF NOT EXISTS t3 (
    a INT AUTO_INCREMENT PRIMARY KEY,
    b INT NULL
);

CREATE TABLE IF NOT EXISTS parent1 (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    CONSTRAINT pk_parent1_id PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS child1 (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    parent_id CHAR(36) NULL,
    CONSTRAINT pk_child1_id PRIMARY KEY (id)
);

-- Testing composite keys
CREATE TABLE IF NOT EXISTS t4 (
    a INT NOT NULL,
    b INT NOT NULL,
    c INT NULL,
    PRIMARY KEY (a, b)
);

CREATE TABLE IF NOT EXISTS t5 (
    x INT NOT NULL,
    y INT NOT NULL,
    z INT NULL
);

-- Testing init schema

-- DELIMITER //
CREATE FUNCTION generate_custom_id()
RETURNS VARCHAR(255)
DETERMINISTIC
BEGIN
    RETURN CONCAT('EMP-', LPAD(FLOOR(RAND() * 1000000), 6, '0'));
END;
-- DELIMITER;

CREATE TABLE IF NOT EXISTS employee_log (
    id INT NOT NULL AUTO_INCREMENT,
    employee_id VARCHAR(255),
    action VARCHAR(10),
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

-- DELIMITER //
CREATE TRIGGER before_insert_employee_log
BEFORE INSERT ON employee_log
FOR EACH ROW
BEGIN
    IF NEW.employee_id IS NULL THEN
        SET NEW.employee_id = generate_custom_id();
    END IF;
END;
-- DELIMITER;

CREATE TABLE IF NOT EXISTS custom_table (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  data JSON,
  status ENUM('Low', 'Medium', 'High') NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_custom_domain CHECK (name REGEXP '^[a-zA-Z]+$')
);

CREATE INDEX idx_name ON custom_table(name);

CREATE TABLE IF NOT EXISTS tablewithcount (
    id VARCHAR(255) NOT NULL
);


-- Foreign key constraints
ALTER TABLE container ADD CONSTRAINT container_container_status_fk FOREIGN KEY (container_status_id) REFERENCES container_status (id);

ALTER TABLE t1 ADD CONSTRAINT t1_b_fkey FOREIGN KEY (b) REFERENCES t1(a);
ALTER TABLE t2 ADD CONSTRAINT t2_b_fkey FOREIGN KEY (b) REFERENCES t3(a);
ALTER TABLE t3 ADD CONSTRAINT t3_b_fkey FOREIGN KEY (b) REFERENCES t2(a);
ALTER TABLE child1 ADD CONSTRAINT fk_child1_parent_id_parent1_id FOREIGN KEY (parent_id) REFERENCES parent1(id) ON DELETE CASCADE;
ALTER TABLE t5 ADD CONSTRAINT t5_t4_fkey FOREIGN KEY (x, y) REFERENCES t4 (a, b);


-- Disable foreign key checks
SET foreign_key_checks = 0;


INSERT INTO container_status (id) VALUES (1), (2), (3), (4), (5);

INSERT INTO container (code, container_status_id) VALUES 
('code1', 1),
('code2', 2),
('code3', 3),
('code4', 4),
('code5', 5);

INSERT INTO users (first_name, last_name, age, current_salary) VALUES
('John', 'Doe', 30, 50000.00),
('Jane', 'Smith', 25, 60000.00),
('Alice', 'Johnson', 28, 55000.00),
('Bob', 'Williams', 35, 65000.00),
('Charlie', 'Brown', 40, 70000.00);

INSERT INTO unique_emails (email) VALUES
('john.doe@example.com'),
('jane.smith@example.com'),
('alice.johnson@example.com'),
('bob.williams@example.com'),
('charlie.brown@example.com');

INSERT INTO unique_emails_and_usernames (email, username) VALUES
('john.doe@example.com', 'johndoe'),
('jane.smith@example.com', 'janesmith'),
('alice.johnson@example.com', 'alicejohnson'),
('bob.williams@example.com', 'bobwilliams'),
('charlie.brown@example.com', 'charliebrown');

INSERT INTO t1 (b) VALUES (NULL), (1), (2), (3), (4);

INSERT INTO t2 (b) VALUES (NULL), (1), (2), (3), (4);

INSERT INTO t3 (b) VALUES (NULL), (1), (2), (3), (4);

INSERT INTO parent1 (id) VALUES 
('550e8400-e29b-41d4-a716-446655440000'), 
('550e8400-e29b-41d4-a716-446655440001'), 
('550e8400-e29b-41d4-a716-446655440002'), 
('550e8400-e29b-41d4-a716-446655440003'), 
('550e8400-e29b-41d4-a716-446655440004');

INSERT INTO child1 (parent_id) VALUES 
(NULL), 
('550e8400-e29b-41d4-a716-446655440000'), 
('550e8400-e29b-41d4-a716-446655440001'), 
('550e8400-e29b-41d4-a716-446655440002'), 
('550e8400-e29b-41d4-a716-446655440003');

INSERT INTO t4 (a, b, c) VALUES 
(1, 1, 100), (2, 2, 200), (3, 3, 300), (4, 4, 400), (5, 5, 500);

INSERT INTO t5 (x, y, z) VALUES 
(1, 1, 10), (2, 2, 20), (3, 3, 30), (4, 4, 40), (5, 5, 50);

INSERT INTO employee_log (employee_id, action) VALUES 
(NULL, 'INSERT'), (NULL, 'UPDATE'), (NULL, 'DELETE'), (NULL, 'INSERT'), (NULL, 'UPDATE');

INSERT INTO custom_table (name, data, status) VALUES 
('Alpha', '{}', 'Low'), ('Beta', '{}', 'Medium'), ('Gamma', '{}', 'High'), ('Delta', '{}', 'Low'), ('Epsilon', '{}', 'Medium');

INSERT INTO tablewithcount (id) VALUES 
('1'), ('2'), ('3'), ('4'), ('5');

-- Re-enable foreign key checks
SET foreign_key_checks = 1;
