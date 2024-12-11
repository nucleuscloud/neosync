EXEC('
CREATE SCHEMA mssqltest;
');

CREATE TABLE mssqltest.users (
    user_id INT IDENTITY(1,1) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    manager_id INT,
    mentor_id INT
);

CREATE TABLE mssqltest.initiatives (
    initiative_id INT IDENTITY(1,1) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    lead_id INT,
    client_id INT
);

CREATE TABLE mssqltest.tasks (
    task_id INT IDENTITY(1,1) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(50),
    initiative_id INT,
    assignee_id INT,
    reviewer_id INT
);

CREATE TABLE mssqltest.skills (
    skill_id INT IDENTITY(1,1) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(100)
);

CREATE TABLE mssqltest.user_skills (
    user_skill_id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT NOT NULL,
    skill_id INT,
    proficiency_level INT CHECK (proficiency_level BETWEEN 1 AND 5)
);

CREATE TABLE mssqltest.comments (
    comment_id INT IDENTITY(1,1) PRIMARY KEY,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT GETDATE(),
    user_id INT NOT NULL,
    task_id INT,
    initiative_id INT,
    parent_comment_id INT
);

CREATE TABLE mssqltest.attachments (
    attachment_id INT IDENTITY(1,1) PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    uploaded_by INT NOT NULL,
    task_id INT NOT NULL,
    initiative_id INT,
    comment_id INT
);


-- Add foreign key constraints
ALTER TABLE mssqltest.users
    ADD CONSTRAINT fk_user_manager FOREIGN KEY (manager_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_user_mentor FOREIGN KEY (mentor_id) REFERENCES mssqltest.users(user_id);

ALTER TABLE mssqltest.initiatives
    ADD CONSTRAINT fk_initiative_lead FOREIGN KEY (lead_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_initiative_client FOREIGN KEY (client_id) REFERENCES mssqltest.users(user_id);

ALTER TABLE mssqltest.tasks
    ADD CONSTRAINT fk_task_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id),
    CONSTRAINT fk_task_assignee FOREIGN KEY (assignee_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_task_reviewer FOREIGN KEY (reviewer_id) REFERENCES mssqltest.users(user_id);

ALTER TABLE mssqltest.user_skills
    ADD CONSTRAINT fk_user_skill_user FOREIGN KEY (user_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_user_skill_skill FOREIGN KEY (skill_id) REFERENCES mssqltest.skills(skill_id);

ALTER TABLE mssqltest.comments
    ADD CONSTRAINT fk_comment_user FOREIGN KEY (user_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_comment_task FOREIGN KEY (task_id) REFERENCES mssqltest.tasks(task_id),
    CONSTRAINT fk_comment_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id),
    CONSTRAINT fk_comment_parent FOREIGN KEY (parent_comment_id) REFERENCES mssqltest.comments(comment_id);

ALTER TABLE mssqltest.attachments
    ADD CONSTRAINT fk_attachment_user FOREIGN KEY (uploaded_by) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_attachment_task FOREIGN KEY (task_id) REFERENCES mssqltest.tasks(task_id),
    CONSTRAINT fk_attachment_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id),
    CONSTRAINT fk_attachment_comment FOREIGN KEY (comment_id) REFERENCES mssqltest.comments(comment_id);


-- Insert data
INSERT INTO mssqltest.users (name, email, manager_id, mentor_id) VALUES
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

INSERT INTO mssqltest.initiatives (name, description, lead_id, client_id) VALUES
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

INSERT INTO mssqltest.tasks (title, description, status, initiative_id, assignee_id, reviewer_id) VALUES
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

INSERT INTO mssqltest.skills (name, category) VALUES
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

INSERT INTO mssqltest.user_skills (user_id, skill_id, proficiency_level) VALUES
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

SET IDENTITY_INSERT mssqltest.comments ON;
INSERT INTO mssqltest.comments (comment_id, content, user_id, task_id, initiative_id, parent_comment_id) VALUES
(1, 'Great progress on the mockups!', 1, 1, 1, NULL),
(2, 'Thanks! I''ve incorporated the feedback from last meeting.', 3, 1, 1, 1),
(3, 'We need to use OAuth for the login system', 2, 2, 2, NULL),
(4, 'Agreed. I''ll update the design docs accordingly.', 4, 2, 2, 3),
(5, 'Data mapping completed, ready for review', 3, 3, 3, NULL),
(6, 'I''ll start the review process today.', 5, 3, 3, 5),
(7, 'AI model showing promising results', 4, 4, 4, NULL),
(8, 'That''s great news! Can we schedule a demo?', 6, 4, 4, 7),
(9, 'Cloud environment is set up and ready', 5, 5, 5, NULL),
(10, 'Excellent. Let''s start the migration process.', 7, 5, 5, 9),
(11, 'Found several critical vulnerabilities', 6, 6, 6, NULL),
(12, 'Can you prioritize them for our next sprint?', 8, 6, 6, 11),
(13, 'Main performance bottleneck identified', 7, 7, 7, NULL),
(14, 'What''s our plan to address it?', 9, 7, 7, 13),
(15, 'User interface designs are approved', 8, 8, 8, NULL),
(16, 'Great job! Let''s move forward with development.', 10, 8, 8, 15),
(17, 'Smart contracts pass initial tests', 9, 9, 9, NULL),
(18, 'Fantastic! We should schedule a security audit next.', 1, 9, 9, 17),
(19, 'Having issues with sensor compatibility', 10, 10, 10, NULL),
(20, 'I can help troubleshoot. Let''s set up a call.', 2, 10, 10, 19),
(21, 'Found bots!', 5, NULL, NULL, NULL);
SET IDENTITY_INSERT mssqltest.comments OFF;


INSERT INTO mssqltest.attachments (file_name, file_path, uploaded_by, task_id, initiative_id, comment_id) VALUES
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
