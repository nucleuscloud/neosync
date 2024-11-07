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


-- init statement tests
CREATE SEQUENCE [mssqlinit].[InvoiceNumberSeq]
    START WITH 10000
    INCREMENT BY 1;

CREATE TABLE mssqlinit.Invoices (
    InvoiceID INT DEFAULT (NEXT VALUE FOR [mssqlinit].[InvoiceNumberSeq]),
    InvoiceNumber VARCHAR(20),
    UpdatedAt DATETIME2 GENERATED ALWAYS AS ROW START,
    ValidTo DATETIME2 GENERATED ALWAYS AS ROW END,
    PERIOD FOR SYSTEM_TIME (UpdatedAt, ValidTo),
    INDEX idx_invoices_invoiceid (InvoiceID)
);

CREATE TABLE mssqlinit.Customers (
    CustomerID INT PRIMARY KEY IDENTITY(1,1),
    FirstName NVARCHAR(50) NOT NULL,
    LastName NVARCHAR(50) NOT NULL,
    Email NVARCHAR(100) NOT NULL UNIQUE,
    Phone NVARCHAR(20),
    FullName AS CONCAT(FirstName, ' ', LastName) PERSISTED,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    INDEX idx_customer_email (Email)
);

CREATE TABLE mssqlinit.Orders (
    OrderID INT PRIMARY KEY IDENTITY(1,1),
    OrderNumber VARCHAR(20), 
    CustomerID INT REFERENCES mssqlinit.Customers(CustomerID),
    OrderDate DATETIME2 DEFAULT GETDATE(),
    TotalAmount DECIMAL(10,2),
    Status NVARCHAR(20) DEFAULT 'Pending',
    INDEX idx_order_customer (CustomerID),
);

CREATE TABLE mssqlinit.Products (
    ProductID INT PRIMARY KEY IDENTITY(1,1),
    Name NVARCHAR(200) NOT NULL,
    Description NVARCHAR(1000),
    Price DECIMAL(10,2) NOT NULL,
    StockQuantity INT NOT NULL,
    ReorderPoint INT NOT NULL,
    LowStockFlag AS CASE 
        WHEN StockQuantity <= ReorderPoint THEN 1 
        ELSE 0 
    END PERSISTED,
    CreatedAt DATETIME2 DEFAULT GETDATE(),
    INDEX idx_product_price (Price)
);

CREATE TABLE mssqlinit.OrderItems (
    OrderItemID INT PRIMARY KEY IDENTITY(1,1),
    OrderID INT REFERENCES mssqlinit.Orders(OrderID),
    ProductID INT REFERENCES mssqlinit.Products(ProductID),
    Quantity INT NOT NULL,
    UnitPrice DECIMAL(10,2) NOT NULL,
    Subtotal AS (Quantity * UnitPrice) PERSISTED,
    INDEX idx_orderitem_order (OrderID),
    INDEX idx_orderitem_product (ProductID)
);



CREATE TYPE [mssqlinit].[EmailAddress] FROM nvarchar(320) NOT NULL;

CREATE TYPE [mssqlinit].[MoneyAmount] FROM DECIMAL(19,4) NOT NULL;

CREATE TYPE [mssqlinit].[OrderDetailType] AS TABLE
(
    OrderLineId INT,
    ProductId INT NOT NULL,
    Quantity INT NOT NULL,
    UnitPrice DECIMAL(18,2) NOT NULL,
    PRIMARY KEY (OrderLineId)
);


CREATE TABLE mssqlinit.[Employee] (
    EmployeeId INT,
    FirstName NVARCHAR(50) NOT NULL,
    LastName NVARCHAR(50) NOT NULL,
    Email NVARCHAR(255) NOT NULL,
    Age INT,
    Salary DECIMAL(18,2),
    Department NVARCHAR(50),
    HireDate DATE,
    Status NVARCHAR(20),
    PhoneNumber NVARCHAR(20),
    ZipCode NVARCHAR(10),
    Gender CHAR(1),
    
    -- Basic range check
    CONSTRAINT CHK_Employee_Age CHECK (Age >= 18 AND Age < 150),
    
    -- Salary validation with multiple conditions
    CONSTRAINT CHK_Employee_Salary CHECK (Salary >= 0 AND Salary <= 1000000),
    
    -- Pattern matching using LIKE
    CONSTRAINT CHK_Employee_Email CHECK (Email LIKE '%_@_%._%'),
    
    -- List of allowed values
    CONSTRAINT CHK_Employee_Department CHECK (Department IN ('IT', 'HR', 'Sales', 'Marketing', 'Finance')),
    
    -- Date validation
    CONSTRAINT CHK_Employee_HireDate CHECK (HireDate >= '2000-01-01' AND HireDate <= GETDATE()),
    
    -- Complex string pattern (US phone format)
    CONSTRAINT CHK_Employee_Phone CHECK (
        PhoneNumber LIKE '[0-9][0-9][0-9]-[0-9][0-9][0-9]-[0-9][0-9][0-9][0-9]' OR
        PhoneNumber LIKE '([0-9][0-9][0-9]) [0-9][0-9][0-9]-[0-9][0-9][0-9][0-9]'
    ),
    
    -- Status with custom message
    CONSTRAINT CHK_Employee_Status CHECK (Status IN ('Active', 'Inactive', 'On Leave', 'Terminated')),
    
    -- ZIP code format (US)
    CONSTRAINT CHK_Employee_ZipCode CHECK (
        ZipCode LIKE '[0-9][0-9][0-9][0-9][0-9]' OR
        ZipCode LIKE '[0-9][0-9][0-9][0-9][0-9]-[0-9][0-9][0-9][0-9]'
    ),
    
    -- Gender validation
    CONSTRAINT CHK_Employee_Gender CHECK (Gender IN ('M', 'F', 'N')),

    CONSTRAINT PK_Employees PRIMARY KEY CLUSTERED (EmployeeId),
    CONSTRAINT UQ_Employees_Email UNIQUE (Email)
);


CREATE NONCLUSTERED COLUMNSTORE INDEX [IX_Employees_Analytics] 
ON [mssqlinit].[Employee] 
(
    [Department],
    [Salary],
    [HireDate],
    [Status]
);
