CREATE TABLE neo_schema.alldatatypes (
    id INT IDENTITY (1, 1) PRIMARY KEY,
    -- Exact numerics
    col_bigint BIGINT,
    col_numeric NUMERIC(18,0),
    col_bit BIT,
    col_smallint SMALLINT,
    col_decimal DECIMAL(18,0),
    col_smallmoney SMALLMONEY,
    col_int INT,
    col_tinyint TINYINT,
    col_money MONEY,

    -- Approximate numerics
    col_float FLOAT,
    col_real REAL,

    -- Date and time
    col_date DATE,
    col_datetimeoffset DATETIMEOFFSET,
    col_datetime2 DATETIME2,
    col_smalldatetime SMALLDATETIME,
    col_datetime DATETIME,
    col_time TIME,

    -- Character strings
    col_char CHAR(10),
    col_varchar VARCHAR(50),
    col_text TEXT,

    -- Unicode character strings
    col_nchar NCHAR(10),
    col_nvarchar NVARCHAR(50),
    col_ntext NTEXT,

    -- Binary strings 
    col_binary BINARY(10),
    col_varbinary VARBINARY(50),
    -- col_image IMAGE, BROKEN

    -- Other data types 
    col_uniqueidentifier UNIQUEIDENTIFIER,
    col_xml XML
    -- BROKEN
    -- col_geography GEOGRAPHY,
    -- col_geometry GEOMETRY,
    -- col_hierarchyid HIERARCHYID,
    -- col_sql_variant SQL_VARIANT
);

-- Temporal Table
CREATE TABLE neo_schema.temporal_table (
    id INT NOT NULL PRIMARY KEY CLUSTERED,
    valid_from DATETIME2 GENERATED ALWAYS AS ROW START,
    valid_to DATETIME2 GENERATED ALWAYS AS ROW END,
    PERIOD FOR SYSTEM_TIME(valid_from, valid_to)
)

INSERT INTO neo_schema.temporal_table (id)
VALUES (2),
       (3),
       (4);


INSERT INTO neo_schema.alldatatypes (
    -- Exact numerics
    col_bigint, col_numeric, 
    col_bit, 
    col_smallint, col_decimal, col_smallmoney, col_int, col_tinyint, col_money,
    -- Approximate numerics
    col_float, col_real,
    -- Date and time
    col_date, col_datetimeoffset, col_datetime2, col_smalldatetime, col_datetime, col_time,
    -- Character strings
    col_char, col_varchar, col_text,
    -- Unicode character strings
    col_nchar, col_nvarchar, col_ntext,
    -- -- Binary strings
    col_binary, col_varbinary, 
    -- col_image,
    -- Other data types
    col_uniqueidentifier,
    col_xml
    -- col_geography, 
    -- col_geometry,
    -- col_hierarchyid, col_sql_variant
)
VALUES (
    -- Exact numerics
    9223372036854775807, -- BIGINT max value
    1234567890, -- NUMERIC
    1, -- BIT
    32767, -- SMALLINT max value
    1234567890, -- DECIMAL
    214748.3647, -- SMALLMONEY max value
    2147483647, -- INT max value
    255, -- TINYINT max value
    922337203685477.5807, -- MONEY max value
    
    -- Approximate numerics
    1234.56789, -- FLOAT
    1234.56, -- REAL
    
    -- Date and time
    '2023-05-15', -- DATE
    '2023-05-15 14:30:00 +02:00', -- DATETIMEOFFSET
    '2023-05-15 14:30:00.1234567', -- DATETIME2
    '2023-05-15 14:30:00', -- SMALLDATETIME
    '2023-05-15 14:30:00.123', -- DATETIME
    '14:30:00.1234567', -- TIME
    
    -- Character strings
    'CHAR      ', -- CHAR(10)
    'VARCHAR', -- VARCHAR(50)
    'This is a TEXT column', -- TEXT
    
    -- Unicode character strings
    N'NCHAR     ', -- NCHAR(10)
    N'NVARCHAR', -- NVARCHAR(50)
    N'This is an NTEXT column', -- NTEXT
    
    -- -- Binary strings
    CAST('NEOSYNC' AS binary(10)), -- BINARY(10)
    CAST('NEOSYNC' AS varbinary(50)), -- VARBINARY(50)
    -- 0x0123456789ABCDEF0123456789ABCDEF, -- IMAGE
    
    -- Other data types
   '707f085c-254e-4237-a9fc-5bf05d4298b8', -- UNIQUEIDENTIFIER
    '<root><element>XML Data</element></root>' -- XML
    -- geography::Point(47.65100, -122.34900, 4326), -- GEOGRAPHY
    -- geometry::STGeomFromText('POINT (3 4)', 0), -- GEOMETRY
    -- '/1/2/3/', -- HIERARCHYID
    -- CAST('2023-05-15' AS SQL_VARIANT) -- SQL_VARIANT
);
