INSERT INTO alltypes.alldatatypes (
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
    col_nchar, col_nvarchar, col_json, col_ntext,
    -- -- Binary strings
    -- col_binary, col_varbinary, col_image,
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
    N'{"key": "value"}', -- JSON
    N'This is an NTEXT column', -- NTEXT
    
    -- -- Binary strings
    -- 0x0123456789, -- BINARY(10)
    -- 0x0123456789ABCDEF, -- VARBINARY(50)
    -- 0x0123456789ABCDEF0123456789ABCDEF, -- IMAGE
    
    -- Other data types
   '123e4567-e89b-12d3-a456-426614174000', -- UNIQUEIDENTIFIER
    '<root><element>XML Data</element></root>' -- XML
    -- geography::Point(47.65100, -122.34900, 4326), -- GEOGRAPHY
    -- geometry::STGeomFromText('POINT (3 4)', 0), -- GEOMETRY
    -- '/1/2/3/', -- HIERARCHYID
    -- CAST('2023-05-15' AS SQL_VARIANT) -- SQL_VARIANT
);
