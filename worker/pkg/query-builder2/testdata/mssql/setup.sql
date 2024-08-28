EXEC('
CREATE SCHEMA mssqltest;
');

CREATE TABLE mssqltest.users (
    user_id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    email NVARCHAR(100) UNIQUE NOT NULL,
    manager_id INT,
    mentor_id INT,
    department_id INT,
    primary_project_id INT
);

CREATE TABLE mssqltest.departments (
    department_id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    manager_id INT,
    parent_department_id INT
);

CREATE TABLE mssqltest.projects (
    project_id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    department_id INT,
    lead_id INT,
    initiative_id INT
);

CREATE TABLE mssqltest.initiatives (
    initiative_id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    description NVARCHAR(MAX),
    lead_id INT,
    client_id INT,
    department_id INT
);

CREATE TABLE mssqltest.tasks (
    task_id INT IDENTITY(1,1) PRIMARY KEY,
    title NVARCHAR(200) NOT NULL,
    description NVARCHAR(MAX),
    status NVARCHAR(50),
    initiative_id INT,
    project_id INT,
    assignee_id INT,
    reviewer_id INT,
    department_id INT
);

CREATE TABLE mssqltest.skills (
    skill_id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    category NVARCHAR(100),
    parent_skill_id INT
);

CREATE TABLE mssqltest.user_skills (
    user_skill_id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT,
    skill_id INT,
    proficiency_level INT CHECK (proficiency_level BETWEEN 1 AND 5)
);

CREATE TABLE mssqltest.comments (
    comment_id INT IDENTITY(1,1) PRIMARY KEY,
    content NVARCHAR(MAX) NOT NULL,
    created_at DATETIME DEFAULT GETDATE(),
    user_id INT,
    task_id INT,
    initiative_id INT,
    project_id INT,
    parent_comment_id INT
);

CREATE TABLE mssqltest.attachments (
    attachment_id INT IDENTITY(1,1) PRIMARY KEY,
    file_name NVARCHAR(255) NOT NULL,
    file_path NVARCHAR(255) NOT NULL,
    uploaded_by INT,
    task_id INT,
    initiative_id INT,
    project_id INT,
    comment_id INT,
    skill_id INT
);

CREATE TABLE mssqltest.employee_projects (
    employee_project_id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT,
    project_id INT,
    role NVARCHAR(100)
);

-- Add foreign key constraints
ALTER TABLE mssqltest.users
    ADD CONSTRAINT fk_user_manager FOREIGN KEY (manager_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_user_mentor FOREIGN KEY (mentor_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_user_department FOREIGN KEY (department_id) REFERENCES mssqltest.departments(department_id),
    CONSTRAINT fk_user_primary_project FOREIGN KEY (primary_project_id) REFERENCES mssqltest.projects(project_id);

ALTER TABLE mssqltest.departments
    ADD CONSTRAINT fk_department_manager FOREIGN KEY (manager_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_department_parent FOREIGN KEY (parent_department_id) REFERENCES mssqltest.departments(department_id);

ALTER TABLE mssqltest.projects
    ADD CONSTRAINT fk_project_department FOREIGN KEY (department_id) REFERENCES mssqltest.departments(department_id),
    CONSTRAINT fk_project_lead FOREIGN KEY (lead_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_project_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id);

ALTER TABLE mssqltest.initiatives
    ADD CONSTRAINT fk_initiative_lead FOREIGN KEY (lead_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_initiative_client FOREIGN KEY (client_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_initiative_department FOREIGN KEY (department_id) REFERENCES mssqltest.departments(department_id);

ALTER TABLE mssqltest.tasks
    ADD CONSTRAINT fk_task_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id),
    CONSTRAINT fk_task_project FOREIGN KEY (project_id) REFERENCES mssqltest.projects(project_id),
    CONSTRAINT fk_task_assignee FOREIGN KEY (assignee_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_task_reviewer FOREIGN KEY (reviewer_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_task_department FOREIGN KEY (department_id) REFERENCES mssqltest.departments(department_id);

ALTER TABLE mssqltest.skills
    ADD CONSTRAINT fk_skill_parent FOREIGN KEY (parent_skill_id) REFERENCES mssqltest.skills(skill_id);

ALTER TABLE mssqltest.user_skills
    ADD CONSTRAINT fk_user_skill_user FOREIGN KEY (user_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_user_skill_skill FOREIGN KEY (skill_id) REFERENCES mssqltest.skills(skill_id);

ALTER TABLE mssqltest.comments
    ADD CONSTRAINT fk_comment_user FOREIGN KEY (user_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_comment_task FOREIGN KEY (task_id) REFERENCES mssqltest.tasks(task_id),
    CONSTRAINT fk_comment_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id),
    CONSTRAINT fk_comment_project FOREIGN KEY (project_id) REFERENCES mssqltest.projects(project_id),
    CONSTRAINT fk_comment_parent FOREIGN KEY (parent_comment_id) REFERENCES mssqltest.comments(comment_id);

ALTER TABLE mssqltest.attachments
    ADD CONSTRAINT fk_attachment_user FOREIGN KEY (uploaded_by) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_attachment_task FOREIGN KEY (task_id) REFERENCES mssqltest.tasks(task_id),
    CONSTRAINT fk_attachment_initiative FOREIGN KEY (initiative_id) REFERENCES mssqltest.initiatives(initiative_id),
    CONSTRAINT fk_attachment_project FOREIGN KEY (project_id) REFERENCES mssqltest.projects(project_id),
    CONSTRAINT fk_attachment_comment FOREIGN KEY (comment_id) REFERENCES mssqltest.comments(comment_id),
    CONSTRAINT fk_attachment_skill FOREIGN KEY (skill_id) REFERENCES mssqltest.skills(skill_id);

ALTER TABLE mssqltest.employee_projects
    ADD CONSTRAINT fk_employee_project_user FOREIGN KEY (user_id) REFERENCES mssqltest.users(user_id),
    CONSTRAINT fk_employee_project_project FOREIGN KEY (project_id) REFERENCES mssqltest.projects(project_id);

-- Sample data insertions
INSERT INTO mssqltest.departments (name, manager_id, parent_department_id) VALUES
('IT', NULL, NULL),
('HR', NULL, NULL),
('Finance', NULL, NULL);

INSERT INTO mssqltest.users (name, email, manager_id, mentor_id, department_id, primary_project_id) VALUES
('John Doe', 'john.doe@example.com', NULL, NULL, 1, NULL),
('Jane Smith', 'jane.smith@example.com', 1, NULL, 2, NULL),
('Bob Johnson', 'bob.johnson@example.com', 1, 2, 1, NULL),
('Alice Williams', 'alice.williams@example.com', 2, 1, 2, NULL),
('Charlie Brown', 'charlie.brown@example.com', 2, 3, 3, NULL);

UPDATE mssqltest.departments SET manager_id = 1 WHERE department_id = 1;
UPDATE mssqltest.departments SET manager_id = 2 WHERE department_id = 2;
UPDATE mssqltest.departments SET manager_id = 3 WHERE department_id = 3;

INSERT INTO mssqltest.initiatives (name, description, lead_id, client_id, department_id) VALUES
('Website Redesign', 'Overhaul company website', 1, 2, 1),
('Mobile App Development', 'Create a new mobile app', 2, 3, 1),
('Data Migration', 'Migrate data to new system', 3, 4, 3),
('AI Integration', 'Implement AI in current products', 4, 5, 1),
('Cloud Migration', 'Move infrastructure to the cloud', 5, 1, 1);

INSERT INTO mssqltest.projects (name, department_id, lead_id, initiative_id) VALUES
('Frontend Redesign', 1, 1, 1),
('Backend Revamp', 1, 3, 1),
('iOS App', 1, 2, 2),
('Android App', 1, 4, 2),
('Database Migration', 3, 5, 3);

UPDATE mssqltest.users SET primary_project_id = 
    CASE 
        WHEN user_id = 1 THEN 1
        WHEN user_id = 2 THEN 3
        WHEN user_id = 3 THEN 2
        WHEN user_id = 4 THEN 4
        WHEN user_id = 5 THEN 5
    END;

INSERT INTO mssqltest.tasks (title, description, status, initiative_id, project_id, assignee_id, reviewer_id, department_id) VALUES
('Design mockups', 'Create initial design mockups', 'In Progress', 1, 1, 3, 1, 1),
('Develop login system', 'Implement secure login system', 'Not Started', 2, 3, 4, 2, 1),
('Data mapping', 'Map data fields between systems', 'Completed', 3, 5, 5, 3, 3),
('Train AI model', 'Train and test initial AI model', 'In Progress', 4, NULL, 1, 4, 1),
('Setup cloud environment', 'Initialize cloud infrastructure', 'In Progress', 5, NULL, 2, 5, 1);

INSERT INTO mssqltest.skills (name, category, parent_skill_id) VALUES
('Programming', 'Technical', NULL),
('JavaScript', 'Programming', 1),
('Python', 'Programming', 1),
('SQL', 'Database', NULL),
('Project Management', 'Management', NULL);

INSERT INTO mssqltest.user_skills (user_id, skill_id, proficiency_level) VALUES
(1, 2, 5),
(2, 3, 4),
(3, 4, 5),
(4, 5, 4),
(5, 2, 3);

INSERT INTO mssqltest.comments (content, user_id, task_id, initiative_id, project_id, parent_comment_id) VALUES
('Great progress on the mockups!', 1, 1, 1, 1, NULL),
('Thanks! I''ve incorporated the feedback from last meeting.', 3, 1, 1, 1, 1),
('We need to use OAuth for the login system', 2, 2, 2, 3, NULL),
('Agreed. I''ll update the design docs accordingly.', 4, 2, 2, 3, 3),
('Data mapping completed, ready for review', 3, 3, 3, 5, NULL);

INSERT INTO mssqltest.attachments (file_name, file_path, uploaded_by, task_id, initiative_id, project_id, comment_id, skill_id) VALUES
('mockup_v1.png', '/files/mockups/mockup_v1.png', 3, 1, 1, 1, 2, NULL),
('login_flow.pdf', '/files/docs/login_flow.pdf', 4, 2, 2, 3, 4, NULL),
('data_mapping.xlsx', '/files/data/data_mapping.xlsx', 5, 3, 3, 5, 5, NULL),
('ai_model_results.ipynb', '/files/notebooks/ai_model_results.ipynb', 1, 4, 4, NULL, NULL, 3),
('cloud_architecture.jpg', '/files/diagrams/cloud_architecture.jpg', 2, 5, 5, NULL, NULL, NULL);

INSERT INTO mssqltest.employee_projects (user_id, project_id, role) VALUES
(1, 1, 'Project Manager'),
(2, 3, 'Team Lead'),
(3, 2, 'Senior Developer'),
(4, 4, 'UI/UX Designer'),
(5, 5, 'Database Administrator');
