CREATE TABLE testdb.sqlmanagermssql3.users (
  id INT IDENTITY(1,1) PRIMARY KEY,
  first_name VARCHAR(50) NOT NULL,
  last_name VARCHAR(50) NOT NULL,
  fullname AS CONCAT(first_name, ' ', last_name) PERSISTED,
  age INT NOT NULL,
  current_salary FLOAT NOT NULL
);
