USE m_db_1;

CREATE TABLE IF NOT EXISTS container_status (
	id int NOT NULL AUTO_INCREMENT,
	PRIMARY KEY (id));

CREATE TABLE IF NOT EXISTS container (
	id int NOT NULL AUTO_INCREMENT,
	code varchar(32) NOT NULL,
	container_status_id int NOT NULL,
PRIMARY KEY (id),
UNIQUE KEY container_code_uniq (code),
KEY container_container_status_fk (container_status_id)) ENGINE = InnoDB AUTO_INCREMENT = 530 DEFAULT CHARSET = utf8mb3;


ALTER TABLE container ADD CONSTRAINT container_container_status_fk FOREIGN KEY (container_status_id) REFERENCES container_status (id);


USE m_db_2;
CREATE TABLE IF NOT EXISTS container_status (
	id int NOT NULL AUTO_INCREMENT,
	PRIMARY KEY (id)) ENGINE = InnoDB AUTO_INCREMENT = 2 DEFAULT CHARSET = utf8mb3;

CREATE TABLE IF NOT EXISTS container (
	id int NOT NULL AUTO_INCREMENT,
	code varchar(32) NOT NULL,
	container_status_id int NOT NULL,
PRIMARY KEY (id),
UNIQUE KEY container_code_uniq (code),
KEY container_container_status_fk (container_status_id)) ENGINE = InnoDB AUTO_INCREMENT = 530 DEFAULT CHARSET = utf8mb3;

ALTER TABLE container ADD CONSTRAINT container_container_status_fk FOREIGN KEY (container_status_id) REFERENCES container_status (id);
