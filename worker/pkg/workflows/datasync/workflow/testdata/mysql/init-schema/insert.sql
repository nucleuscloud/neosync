-- Disable foreign key checks
SET foreign_key_checks = 0;

-- Insert data into init_schema
USE init_schema;

INSERT INTO container_status (id) VALUES (1), (2), (3), (4), (5);

INSERT INTO container (code, container_status_id) VALUES 
('code1', 1),
('code2', 2),
('code3', 3),
('code4', 4),
('code5', 5);

-- Insert data into init_schema2
USE init_schema2;

INSERT INTO container_status (id) VALUES (1), (2), (3), (4), (5);

INSERT INTO container (code, container_status_id) VALUES 
('code1', 1),
('code2', 2),
('code3', 3),
('code4', 4),
('code5', 5);

-- Insert data into init_schema3
USE init_schema3;

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
