CREATE TABLE alltypes.alldatatypes (
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
    col_json NVARCHAR(MAX),
    col_ntext NTEXT,

    -- Binary strings BROKEN
    -- col_binary BINARY(10),
    -- col_varbinary VARBINARY(50),
    -- col_image IMAGE,

    -- Other data types 
    col_uniqueidentifier UNIQUEIDENTIFIER,
    col_xml XML
    -- BROKEN
    -- col_geography GEOGRAPHY,
    -- col_geometry GEOMETRY,
    -- col_hierarchyid HIERARCHYID,
    -- col_sql_variant SQL_VARIANT
);
