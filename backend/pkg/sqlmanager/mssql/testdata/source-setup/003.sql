CREATE TABLE testdb.sqlmanagermssql3.users (
  id INT IDENTITY(1,1) PRIMARY KEY,
  first_name VARCHAR(50) NOT NULL,
  last_name VARCHAR(50) NOT NULL,
  fullname AS CONCAT(first_name, ' ', last_name) PERSISTED,
  age INT NOT NULL,
  current_salary FLOAT NOT NULL
);

CREATE TABLE testdb.sqlmanagermssql2.parent1 (
    id1 uniqueidentifier NOT NULL DEFAULT NEWID(),
    id2 INT NOT NULL,
    name NVARCHAR(100),
    CONSTRAINT pk_parent1 PRIMARY KEY (id1, id2)
);

CREATE TABLE testdb.sqlmanagermssql2.child1 (
    id uniqueidentifier NOT NULL DEFAULT NEWID(),
    parent_id1 uniqueidentifier NULL,
    parent_id2 INT NULL,
    description NVARCHAR(200),
    CONSTRAINT pk_child1_id PRIMARY KEY (id),
    CONSTRAINT fk_child1_parent FOREIGN KEY (parent_id1, parent_id2)
        REFERENCES testdb.sqlmanagermssql2.parent1(id1, id2) ON DELETE CASCADE
);

-- Compose Circular Dependencies TableA, TableB

-- Create Table A
CREATE TABLE testdb.sqlmanagermssql2.TableA (
    IdA1 INT NOT NULL,
    IdA2 VARCHAR(10) NOT NULL,
    IdB1 INT NULL,
    IdB2 VARCHAR(10) NULL,
    DataA VARCHAR(100),
    CONSTRAINT PK_TableA PRIMARY KEY (IdA1, IdA2)
);

-- Create TableB
CREATE TABLE testdb.sqlmanagermssql2.TableB (
    IdB1 INT NOT NULL,
    IdB2 VARCHAR(10) NOT NULL,
    IdA1 INT NOT NULL,
    IdA2 VARCHAR(10) NOT NULL,
    DataB VARCHAR(100),
    CONSTRAINT PK_TableB PRIMARY KEY (IdB1, IdB2)
);

-- Add foreign key from TableA to TableB
ALTER TABLE testdb.sqlmanagermssql2.TableA
ADD CONSTRAINT FK_TableA_TableB
FOREIGN KEY (IdB1, IdB2) REFERENCES testdb.sqlmanagermssql2.TableB(IdB1, IdB2);

-- Add foreign key from TableB to TableA
ALTER TABLE testdb.sqlmanagermssql2.TableB
ADD CONSTRAINT FK_TableB_TableA
FOREIGN KEY (IdA1, IdA2) REFERENCES testdb.sqlmanagermssql2.TableA(IdA1, IdA2);


CREATE TABLE testdb.sqlmanagermssql2.defaults_table (
    id INT IDENTITY(1,1) PRIMARY KEY,
    description NVARCHAR(MAX),
    age INT DEFAULT 18,
    is_active BIT DEFAULT 1,
    registration_date DATE DEFAULT GETDATE(),
    last_login DATETIME2,
    score DECIMAL(10,2) DEFAULT 0.00,
    status NVARCHAR(20) DEFAULT 'pending',
    created_at DATETIME2 DEFAULT SYSDATETIME(),
    uuid UNIQUEIDENTIFIER DEFAULT NEWID()
);
