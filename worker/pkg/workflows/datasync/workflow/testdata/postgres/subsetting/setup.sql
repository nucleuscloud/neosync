CREATE SCHEMA IF NOT EXISTS subsetting;

SET search_path TO subsetting;

-- Test_BuildQueryMap_DoubleRootSubset
CREATE TABLE test_2_x (
  id BIGINT NOT NULL PRIMARY KEY,
  name text,
  created timestamp without time zone
);

CREATE TABLE test_2_b (
  id BIGINT NOT NULL PRIMARY KEY,
  name text,
  created timestamp without time zone
);

CREATE TABLE test_2_a (
  id BIGINT NOT NULL PRIMARY KEY,
  x_id BIGINT NOT NULL,
  CONSTRAINT test2_x FOREIGN KEY (x_id) REFERENCES test_2_x (id) ON DELETE CASCADE
);

CREATE TABLE test_2_c (
  id BIGINT NOT NULL PRIMARY KEY,
  name text,
  created timestamp without time zone,
  a_id BIGINT NOT NULL,
  b_id BIGINT NOT NULL,
  CONSTRAINT test2_a FOREIGN KEY (a_id) REFERENCES test_2_a (id) ON DELETE CASCADE,
  CONSTRAINT test2_b FOREIGN KEY (b_id) REFERENCES test_2_b (id) ON DELETE CASCADE
);

CREATE TABLE test_2_d (
  id BIGINT NOT NULL PRIMARY KEY,
  c_id BIGINT NOT NULL,
  CONSTRAINT test2_x FOREIGN KEY (c_id) REFERENCES test_2_c (id) ON DELETE CASCADE
);


CREATE TABLE test_2_e (
  id BIGINT NOT NULL PRIMARY KEY,
  c_id BIGINT NOT NULL,
  CONSTRAINT test2_x FOREIGN KEY (c_id) REFERENCES test_2_c (id) ON DELETE CASCADE
);

INSERT INTO test_2_x (id, name, created) VALUES 
(1, 'Xander', '2023-06-01 10:00:00'),
(2, 'Xena', '2023-06-02 11:00:00'),
(3, 'Xavier', '2023-06-03 12:00:00'),
(4, 'Xiomara', '2023-06-04 13:00:00'),
(5, 'Xaviera', '2023-06-05 14:00:00');

INSERT INTO test_2_b (id, name, created) VALUES 
(1, 'Beta1', '2023-06-01 15:00:00'),
(2, 'Beta2', '2023-06-02 16:00:00'),
(3, 'Beta3', '2023-06-03 17:00:00'),
(4, 'Beta4', '2023-06-04 18:00:00'),
(5, 'Beta5', '2023-06-05 19:00:00');

INSERT INTO test_2_a (id, x_id) VALUES 
(1, 5),
(2, 4),
(3, 3),
(4, 3),
(5, 1);

INSERT INTO test_2_c (id, name, created, a_id, b_id) VALUES 
(1, 'Gamma1', '2023-06-01 20:00:00', 1, 1),
(2, 'Gamma2', '2023-06-02 21:00:00', 2, 2),
(3, 'Gamma3', '2023-06-03 22:00:00', 3, 3),
(4, 'Gamma4', '2023-06-04 23:00:00', 4, 4),
(5, 'Gamma5', '2023-06-05 00:00:00', 5, 5);

INSERT INTO test_2_d (id, c_id) VALUES 
(1, 1),
(2, 2),
(3, 3),
(4, 4),
(5, 5);

INSERT INTO test_2_e (id, c_id) VALUES 
(1, 1),
(2, 2),
(3, 3),
(4, 4),
(5, 5);

-- Test_BuildQueryMap_MultipleRoots, Test_BuildQueryMap_MultipleSubset, Test_BuildQueryMap_MultipleSubsets_SubsetsByForeignKeysOff
CREATE TABLE test_3_a (
  id BIGINT NOT NULL PRIMARY KEY
);
CREATE TABLE test_3_b (
  id BIGINT NOT NULL PRIMARY KEY,
  a_id BIGINT NOT NULL,
  CONSTRAINT test3_a FOREIGN KEY (a_id) REFERENCES test_3_a (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_c (
  id BIGINT NOT NULL PRIMARY KEY,
  b_id BIGINT NOT NULL,
  CONSTRAINT test3_b FOREIGN KEY (b_id) REFERENCES test_3_b (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_d (
  id BIGINT NOT NULL PRIMARY KEY,
  c_id BIGINT NOT NULL,
  CONSTRAINT test3_c FOREIGN KEY (c_id) REFERENCES test_3_c (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_e (
  id BIGINT NOT NULL PRIMARY KEY,
  d_id BIGINT NOT NULL,
  CONSTRAINT test3_d FOREIGN KEY (d_id) REFERENCES test_3_d (id) ON DELETE CASCADE
);
CREATE TABLE test_3_f (
  id BIGINT NOT NULL PRIMARY KEY
);
CREATE TABLE test_3_g (
  id BIGINT NOT NULL PRIMARY KEY,
  f_id BIGINT NOT NULL,
  CONSTRAINT test3_f FOREIGN KEY (f_id) REFERENCES test_3_f (id) ON DELETE CASCADE
);
 CREATE TABLE test_3_h (
  id BIGINT NOT NULL PRIMARY KEY,
  g_id BIGINT NOT NULL,
  CONSTRAINT test3_g FOREIGN KEY (g_id) REFERENCES test_3_g (id) ON DELETE CASCADE
);
CREATE TABLE test_3_i (
  id BIGINT NOT NULL PRIMARY KEY,
  h_id BIGINT NOT NULL,
  CONSTRAINT test3_h FOREIGN KEY (h_id) REFERENCES test_3_h (id) ON DELETE CASCADE
);

INSERT INTO test_3_a (id) VALUES 
(1), 
(2), 
(3), 
(4), 
(5);
INSERT INTO test_3_b (id, a_id) VALUES 
(1, 3), 
(2, 5), 
(3, 1), 
(4, 4), 
(5, 2);
INSERT INTO test_3_c (id, b_id) VALUES 
(1, 2), 
(2, 4), 
(3, 1), 
(4, 3), 
(5, 5);
INSERT INTO test_3_d (id, c_id) VALUES 
(1, 5), 
(2, 1), 
(3, 4), 
(4, 2), 
(5, 3);
INSERT INTO test_3_e (id, d_id) VALUES 
(1, 2), 
(2, 4), 
(3, 1), 
(4, 5), 
(5, 3);
INSERT INTO test_3_f (id) VALUES 
(1), 
(2), 
(3), 
(4), 
(5);
INSERT INTO test_3_g (id, f_id) VALUES 
(1, 5), 
(2, 1), 
(3, 4), 
(4, 2), 
(5, 3);
INSERT INTO test_3_h (id, g_id) VALUES 
(1, 4), 
(2, 2), 
(3, 5), 
(4, 1), 
(5, 3);
INSERT INTO test_3_i (id,h_id) VALUES 
(1, 4), 
(8, 2), 
(3, 3), 
(9, 1), 
(5, 3);

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


-- composite keys
CREATE TABLE division (
    id BIGINT PRIMARY KEY,
    division_name VARCHAR(100),
    location VARCHAR(100)
);
CREATE TABLE employees (
    id BIGINT,
    division_id BIGINT,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(100),
    PRIMARY KEY (id, division_id),
    FOREIGN KEY (division_id) REFERENCES division (id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);
CREATE TABLE projects (
    id BIGINT PRIMARY KEY,
    project_name VARCHAR(100),
    start_date DATE,
    end_date DATE,
    responsible_employee_id BIGINT,
    responsible_division_id BIGINT,
    CONSTRAINT FK_Projects_Employees FOREIGN KEY (responsible_employee_id, responsible_division_id)
    REFERENCES employees (id, division_id)
    ON DELETE SET NULL
    ON UPDATE CASCADE
);

INSERT INTO division (id, division_name, location) VALUES
(1, 'Marketing', 'New York'),
(2, 'Finance', 'London'),
(3, 'Human Resources', 'San Francisco'),
(4, 'IT', 'Berlin'),
(5, 'Customer Service', 'Tokyo');


INSERT INTO employees (id, division_id, first_name, last_name, email) VALUES
(6, 1, 'Alice', 'Johnson', 'alice.johnson@example.com'),
(7, 2, 'Bob', 'Smith', 'bob.smith@example.com'),
(8, 3, 'Carol', 'Martinez', 'carol.martinez@example.com'),
(9, 4, 'David', 'Lee', 'david.lee@example.com'),
(10, 5, 'Eva', 'Kim', 'eva.kim@example.com');


INSERT INTO projects (id, project_name, start_date, end_date, responsible_employee_id, responsible_division_id) VALUES
(11, 'Website Redesign', '2023-05-01', '2023-10-01', 6, 1),
(12, 'Financial Audit', '2023-06-15', '2023-07-15', 7, 2),
(13, 'Hiring Initiative', '2023-09-01', '2024-01-31', 8, 3),
(14, 'Software Development', '2023-05-20', '2023-12-20', 9, 4),
(15, 'Customer Feedback Analysis', '2023-07-01', '2023-11-30', 10, 5);


-- self referencing 
CREATE TABLE bosses (
	id BIGINT PRIMARY KEY,
	manager_id BIGINT,
	big_boss_id BIGINT,
	FOREIGN KEY (manager_id) REFERENCES bosses (id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (big_boss_id) REFERENCES bosses (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE minions (
	id BIGINT PRIMARY KEY,
	boss_id BIGINT,
	FOREIGN KEY (boss_id) REFERENCES bosses (id) ON UPDATE CASCADE ON DELETE CASCADE
);

INSERT INTO bosses (id, manager_id, big_boss_id) VALUES 
(1, NULL, NULL), 
(2, 1, NULL), 
(3, 2, 1), 
(4, 3, 2), 
(5, 4, 3),
(6,NULL,NULL);

INSERT INTO minions (id, boss_id) VALUES 
(1, 4), 
(2, 3), 
(3, 1), 
(4, 5), 
(5, 2);


-- complex case, self referencing, double reference, circular dependency
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    manager_id INTEGER,
    mentor_id INTEGER
);

CREATE TABLE initiatives (
    initiative_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    lead_id INTEGER,
    client_id INTEGER
);

CREATE TABLE tasks (
    task_id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(50),
    initiative_id INTEGER,
    assignee_id INTEGER,
    reviewer_id INTEGER
);

CREATE TABLE skills (
    skill_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(100)
);

CREATE TABLE user_skills (
    user_skill_id SERIAL PRIMARY KEY,
    user_id INTEGER,
    skill_id INTEGER,
    proficiency_level INTEGER CHECK (proficiency_level BETWEEN 1 AND 5)
);

CREATE TABLE comments (
    comment_id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER,
    task_id INTEGER,
    initiative_id INTEGER,
    parent_comment_id INTEGER
);

CREATE TABLE attachments (
    attachment_id SERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    uploaded_by INTEGER,
    task_id INTEGER,
    initiative_id INTEGER,
    comment_id INTEGER
);

ALTER TABLE users
    ADD CONSTRAINT fk_user_manager FOREIGN KEY (manager_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_user_mentor FOREIGN KEY (mentor_id) REFERENCES users(user_id);

ALTER TABLE initiatives
    ADD CONSTRAINT fk_initiative_lead FOREIGN KEY (lead_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_initiative_client FOREIGN KEY (client_id) REFERENCES users(user_id);

ALTER TABLE tasks
    ADD CONSTRAINT fk_task_initiative FOREIGN KEY (initiative_id) REFERENCES initiatives(initiative_id),
    ADD CONSTRAINT fk_task_assignee FOREIGN KEY (assignee_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_task_reviewer FOREIGN KEY (reviewer_id) REFERENCES users(user_id);

ALTER TABLE user_skills
    ADD CONSTRAINT fk_user_skill_user FOREIGN KEY (user_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_user_skill_skill FOREIGN KEY (skill_id) REFERENCES skills(skill_id);

ALTER TABLE comments
    ADD CONSTRAINT fk_comment_user FOREIGN KEY (user_id) REFERENCES users(user_id),
    ADD CONSTRAINT fk_comment_task FOREIGN KEY (task_id) REFERENCES tasks(task_id),
    ADD CONSTRAINT fk_comment_initiative FOREIGN KEY (initiative_id) REFERENCES initiatives(initiative_id),
    ADD CONSTRAINT fk_comment_parent FOREIGN KEY (parent_comment_id) REFERENCES comments(comment_id);

ALTER TABLE attachments
    ADD CONSTRAINT fk_attachment_user FOREIGN KEY (uploaded_by) REFERENCES users(user_id),
    ADD CONSTRAINT fk_attachment_task FOREIGN KEY (task_id) REFERENCES tasks(task_id),
    ADD CONSTRAINT fk_attachment_initiative FOREIGN KEY (initiative_id) REFERENCES initiatives(initiative_id),
    ADD CONSTRAINT fk_attachment_comment FOREIGN KEY (comment_id) REFERENCES comments(comment_id);


INSERT INTO users (name, email, manager_id, mentor_id) VALUES
('John Doe', 'john.doe@example.com', NULL, NULL),
('Jane Smith', 'jane.smith@example.com', 1, NULL),
('Bob Johnson', 'bob.johnson@example.com', 1, 2),
('Alice Williams', 'alice.williams@example.com', 2, 1),
('Charlie Brown', 'charlie.brown@example.com', 2, 3),
('Diana Prince', 'diana.prince@example.com', 3, 4),
('Ethan Hunt', 'ethan.hunt@example.com', 3, 1),
('Fiona Gallagher', 'fiona.gallagher@example.com', 4, 2),
('George Lucas', 'george.lucas@example.com', 4, 5),
('Hannah Montana', 'hannah.montana@example.com', 5, 3);

INSERT INTO initiatives (name, description, lead_id, client_id) VALUES
('Website Redesign', 'Overhaul company website', 1, 2),
('Mobile App Development', 'Create a new mobile app', 2, 3),
('Data Migration', 'Migrate data to new system', 3, 4),
('AI Integration', 'Implement AI in current products', 4, 5),
('Cloud Migration', 'Move infrastructure to the cloud', 5, 6),
('Security Audit', 'Perform comprehensive security audit', 6, 7),
('Performance Optimization', 'Optimize system performance', 7, 8),
('Customer Portal', 'Develop a new customer portal', 8, 9),
('Blockchain Implementation', 'Implement blockchain technology', 9, 10),
('IoT Platform', 'Develop an IoT management platform', 10, 1);

INSERT INTO tasks (title, description, status, initiative_id, assignee_id, reviewer_id) VALUES
('Design mockups', 'Create initial design mockups', 'In Progress', 1, 3, 1),
('Develop login system', 'Implement secure login system', 'Not Started', 2, 4, 2),
('Data mapping', 'Map data fields between systems', 'Completed', 3, 5, 3),
('Train AI model', 'Train and test initial AI model', 'In Progress', 4, 6, 4),
('Setup cloud environment', 'Initialize cloud infrastructure', 'In Progress', 5, 7, 5),
('Vulnerability assessment', 'Identify system vulnerabilities', 'Not Started', 6, 8, 6),
('Code profiling', 'Profile code for performance bottlenecks', 'In Progress', 7, 9, 7),
('Design user interface', 'Design intuitive user interface', 'Completed', 8, 10, 8),
('Smart contract development', 'Develop initial smart contracts', 'In Progress', 9, 1, 9),
('Sensor integration', 'Integrate IoT sensors with platform', 'Not Started', 10, 2, 10);

INSERT INTO skills (name, category) VALUES
('JavaScript', 'Programming'),
('Python', 'Programming'),
('SQL', 'Database'),
('Project Management', 'Management'),
('UI/UX Design', 'Design'),
('Machine Learning', 'Data Science'),
('Network Security', 'Security'),
('Cloud Architecture', 'Infrastructure'),
('Blockchain', 'Technology'),
('IoT', 'Technology');

INSERT INTO user_skills (user_id, skill_id, proficiency_level) VALUES
(1, 1, 5),
(2, 2, 4),
(3, 3, 5),
(4, 4, 4),
(5, 5, 3),
(6, 6, 5),
(7, 7, 4),
(8, 8, 3),
(9, 9, 4),
(10, 10, 5);

INSERT INTO comments (comment_id, content, user_id, task_id, initiative_id, parent_comment_id) VALUES
(1, 'Great progress on the mockups!', 1, 1, 1, NULL),
(2, 'Thanks! Ive incorporated the feedback from last meeting.', 3, 1, 1, 1),
(3, 'We need to use OAuth for the login system', 2, 2, 2, NULL),
(4, 'Agreed. Ill update the design docs accordingly.', 4, 2, 2, 3),
(5, 'Data mapping completed, ready for review', 3, 3, 3, NULL),
(6, 'Ill start the review process today.', 5, 3, 3, 5),
(7, 'AI model showing promising results', 4, 4, 4, NULL),
(8, 'Thats great news! Can we schedule a demo?', 6, 4, 4, 7),
(9, 'Cloud environment is set up and ready', 5, 5, 5, NULL),
(10, 'Excellent. Lets start the migration process.', 7, 5, 5, 9),
(11, 'Found several critical vulnerabilities', 6, 6, 6, NULL),
(12, 'Can you prioritize them for our next sprint?', 8, 6, 6, 11),
(13, 'Main performance bottleneck identified', 7, 7, 7, NULL),
(14, 'Whats our plan to address it?', 9, 7, 7, 13),
(15, 'User interface designs are approved', 8, 8, 8, NULL),
(16, 'Great job! Lets move forward with development.', 10, 8, 8, 15),
(17, 'Smart contracts pass initial tests', 9, 9, 9, NULL),
(18, 'Fantastic! We should schedule a security audit next.', 1, 9, 9, 17),
(19, 'Having issues with sensor compatibility', 10, 10, 10, NULL),
(20, 'I can help troubleshoot. Lets set up a call.', 2, 10, 10, 19);

INSERT INTO attachments (file_name, file_path, uploaded_by, task_id, initiative_id, comment_id) VALUES
('mockup_v1.png', '/files/mockups/mockup_v1.png', 3, 1, 1, 2),
('login_flow.pdf', '/files/docs/login_flow.pdf', 4, 2, 2, 4),
('data_mapping.xlsx', '/files/data/data_mapping.xlsx', 5, 3, 3, 5),
('ai_model_results.ipynb', '/files/notebooks/ai_model_results.ipynb', 6, 4, 4, 7),
('cloud_architecture.jpg', '/files/diagrams/cloud_architecture.jpg', 7, 5, 5, 9),
('security_report.pdf', '/files/reports/security_report.pdf', 8, 6, 6, 11),
('performance_analysis.html', '/files/reports/performance_analysis.html', 9, 7, 7, 13),
('ui_designs.sketch', '/files/designs/ui_designs.sketch', 10, 8, 8, 15),
('smart_contracts.sol', '/files/blockchain/smart_contracts.sol', 1, 9, 9, 17),
('sensor_specs.pdf', '/files/iot/sensor_specs.pdf', 2, 10, 10, 19);



CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    blueprint_id INTEGER NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blueprints (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    account_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO accounts (name, email, blueprint_id)
VALUES 
    ('John Doe', 'john@example.com', 5),
    ('Jane Smith', 'jane@example.com', NULL),
    ('Bob Wilson', 'bob@example.com', 2),
    ('Alice Brown', 'alice@example.com', 3),
    ('Charlie Davis', 'charlie@example.com', 4)
RETURNING id;

-- Then insert blueprints with account_id references
INSERT INTO blueprints (name, description, account_id)
VALUES 
    ('Basic blueprint', 'A simple starter blueprint', 1),
    ('Pro blueprint', 'Advanced features blueprint', 2),
    ('Team blueprint', 'Collaborative workspace blueprint', 3),
    ('Enterprise blueprint', 'Full-featured business blueprint', 4),
    ('Custom blueprint', 'Customizable blueprint', 5)
RETURNING id;

ALTER TABLE accounts 
ADD CONSTRAINT fk_accounts_blueprints 
FOREIGN KEY (blueprint_id) 
REFERENCES blueprints(id);

ALTER TABLE blueprints 
ADD CONSTRAINT fk_blueprints_accounts 
FOREIGN KEY (account_id) 
REFERENCES accounts(id);


CREATE TABLE IF NOT EXISTS clients (
  id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS titles (
  id SERIAL PRIMARY KEY,
  account_id INTEGER NOT NULL,
  FOREIGN KEY (account_id) REFERENCES accounts (id)
);

CREATE TABLE IF NOT EXISTS client_account_titles (
  id SERIAL PRIMARY KEY,
  account_id INTEGER NOT NULL,
  title_id INTEGER NULL,
  client_id INTEGER NOT NULL,
  FOREIGN KEY (account_id) REFERENCES accounts(id),
  FOREIGN KEY (title_id) REFERENCES titles(id),
  FOREIGN KEY (client_id) REFERENCES clients(id)
);

INSERT INTO clients (id)
VALUES 
    (1),
    (2),
    (3),
    (4),
    (5),
    (6),
    (7),
    (8),
    (9);

-- Insert titles with references to existing accounts
INSERT INTO titles (id, account_id)
VALUES 
    (1, 1),  -- Title for John Doe's account
    (2, 1),  -- Another title for John Doe's account
    (3, 2),  -- Title for Jane Smith's account
    (4, 3),  -- Title for Bob Wilson's account
    (5, 4),  -- Title for Alice Brown's account
    (6, 5);  -- Title for Charlie Davis's account

-- Insert client_account_titles relationships
INSERT INTO client_account_titles (account_id, title_id, client_id)
VALUES
    -- John Doe's account relationships
    (1, 1, 1),  -- First client with first title
    (1, 2, 2),  -- Second client with second title
    (1, null, 9),  -- Third client with null title
    
    -- Jane Smith's account relationships
    (2, 3, 3),  -- Third client
    
    -- Bob Wilson's account relationships
    (3, 4, 4),  -- Fourth client
    
    -- Alice Brown's account relationships
    (4, 5, 5),  -- Fifth client
    
    -- Charlie Davis's account relationships
    (5, 6, 6),  -- Sixth client
    (5, 6, 7),  -- Seventh client sharing same title as sixth client
    (5, null, 8);
