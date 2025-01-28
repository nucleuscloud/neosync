CREATE DATABASE IF NOT EXISTS sqlmanagermysql;
CREATE DATABASE IF NOT EXISTS sqlmanagermysql2;
CREATE DATABASE IF NOT EXISTS sqlmanagermysql3;
CREATE DATABASE IF NOT EXISTS sqlmanagermysql4;

CREATE TABLE IF NOT EXISTS `sqlmanagermysql`.`container_status` (
	`id` int NOT NULL AUTO_INCREMENT,
	PRIMARY KEY (`id`)) ENGINE = InnoDB AUTO_INCREMENT = 2 DEFAULT CHARSET = utf8mb3;

CREATE TABLE IF NOT EXISTS `sqlmanagermysql`.`container` (
	`id` int NOT NULL AUTO_INCREMENT,
	`code` varchar(32) NOT NULL,
	`container_status_id` int NOT NULL,
PRIMARY KEY (`id`),
UNIQUE KEY `container_code_uniq` (`code`),
KEY `container_container_status_fk` (`container_status_id`),
CONSTRAINT `container_container_status_fk` FOREIGN KEY (`container_status_id`) REFERENCES `container_status` (`id`)) ENGINE = InnoDB AUTO_INCREMENT = 530 DEFAULT CHARSET = utf8mb3;


CREATE TABLE IF NOT EXISTS `sqlmanagermysql2`.`container_status` (
	`id` int NOT NULL AUTO_INCREMENT,
	PRIMARY KEY (`id`)) ENGINE = InnoDB AUTO_INCREMENT = 2 DEFAULT CHARSET = utf8mb3;

CREATE TABLE IF NOT EXISTS `sqlmanagermysql2`.`container` (
	`id` int NOT NULL AUTO_INCREMENT,
	`code` varchar(32) NOT NULL,
	`container_status_id` int NOT NULL,
PRIMARY KEY (`id`),
UNIQUE KEY `container_code_uniq` (`code`),
KEY `container_container_status_fk` (`container_status_id`),
CONSTRAINT `container_container_status_fk` FOREIGN KEY (`container_status_id`) REFERENCES `container_status` (`id`)) ENGINE = InnoDB AUTO_INCREMENT = 530 DEFAULT CHARSET = utf8mb3;


USE sqlmanagermysql3;

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    fullname varchar(101) GENERATED ALWAYS AS (CONCAT(first_name,' ',last_name)),
    age int NOT NULL,
    current_salary float NOT NULL
);

CREATE TABLE unique_emails (
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
    b INT NULL,
    CONSTRAINT t1_b_fkey FOREIGN KEY (b) REFERENCES t1(a)
);

CREATE TABLE IF NOT EXISTS t2 (
    a INT AUTO_INCREMENT PRIMARY KEY,
    b INT NULL
);

CREATE TABLE IF NOT EXISTS t3 (
    a INT AUTO_INCREMENT PRIMARY KEY,
    b INT NULL
);

ALTER TABLE t2
ADD CONSTRAINT t2_b_fkey FOREIGN KEY (b) REFERENCES t3(a);

ALTER TABLE t3
ADD CONSTRAINT t3_b_fkey FOREIGN KEY (b) REFERENCES t2(a);


CREATE TABLE IF NOT EXISTS parent1 (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    CONSTRAINT pk_parent1_id PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS child1 (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    parent_id CHAR(36) NULL,
    CONSTRAINT pk_child1_id PRIMARY KEY (id),
    CONSTRAINT fk_child1_parent_id_parent1_id FOREIGN KEY (parent_id) REFERENCES parent1(id) ON DELETE CASCADE
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
    z INT NULL,
    CONSTRAINT t5_t4_fkey FOREIGN KEY (x, y) REFERENCES t4 (a, b)
);



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

INSERT INTO tablewithcount(id) VALUES ('1'), ('2');

-- Creates table that uses reserved MySQL words
CREATE TABLE `order` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `select` VARCHAR(255) NOT NULL,
  `from` DATE NOT NULL,
  `where` VARCHAR(100),
  `group` DECIMAL(10, 2),
  `desc` TEXT
);

-- Creates an index that uses reserved MySQL words
CREATE INDEX `order_index_on_reserved_words` ON `order` (`select`, `from`, `where`);

-- sqlmanagermysql4

USE sqlmanagermysql4;


CREATE TABLE parent1 (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    CONSTRAINT pk_parent1_id PRIMARY KEY (id)
);

CREATE TABLE child1 (
    id CHAR(36) NOT NULL DEFAULT (UUID()),
    parent_id CHAR(36) NULL,
    CONSTRAINT pk_child1_id PRIMARY KEY (id),
    CONSTRAINT fk_child1_parent_id_parent1_id FOREIGN KEY (parent_id) REFERENCES parent1(id) ON DELETE CASCADE
);

-- Creates table that uses reserved MySQL words
CREATE TABLE `order` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `select` VARCHAR(255) NOT NULL,
  `from` DATE NOT NULL,
  `where` VARCHAR(100),
  `group` DECIMAL(10, 2),
  `desc` TEXT
);

-- Creates an index that uses reserved MySQL words
CREATE INDEX `order_index_on_reserved_words` ON `order` (`select`, `from`, `where`);

-- Create a table with some columns
CREATE TABLE test_mixed_index (
    id INT PRIMARY KEY,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    birth_date DATE
);

-- Create a composite index that uses both regular columns and expressions
CREATE INDEX idx_mixed ON test_mixed_index
    (first_name, (UPPER(last_name)), birth_date, (YEAR(birth_date)));
