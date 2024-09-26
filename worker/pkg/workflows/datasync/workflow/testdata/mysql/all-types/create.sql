CREATE DATABASE IF NOT EXISTS all_types;

USE all_types; 
CREATE TABLE IF NOT EXISTS  all_data_types (
    -- Auto-incrementing primary key
    id INT AUTO_INCREMENT PRIMARY KEY,

    -- Numeric Types
    tinyint_col TINYINT,
    smallint_col SMALLINT,
    mediumint_col MEDIUMINT,
    int_col INT,
    bigint_col BIGINT,
    decimal_col DECIMAL(10, 2),
    float_col FLOAT(7, 4),
    double_col DOUBLE(15, 8),
    bit_col BIT(8),

    -- Date and Time Types
    date_col DATE,
    time_col TIME,
    datetime_col DATETIME,
    timestamp_col TIMESTAMP,
    year_col YEAR,

    -- String Types
    char_col CHAR(10),
    varchar_col VARCHAR(255),
    binary_col BINARY(3),
    varbinary_col VARBINARY(255),
    tinyblob_col TINYBLOB,
    tinytext_col TINYTEXT,
    blob_col BLOB,
    text_col TEXT,
    mediumblob_col MEDIUMBLOB,
    mediumtext_col MEDIUMTEXT,
    longblob_col LONGBLOB,
    longtext_col LONGTEXT,
    enum_col ENUM('value1', 'value2', 'value3'),
    set_col SET('option1', 'option2', 'option3'),

    -- Spatial Data Types  BROKEN
    -- geometry_col GEOMETRY,
    -- point_col POINT,
    -- linestring_col LINESTRING,
    -- polygon_col POLYGON,
    -- multipoint_col MULTIPOINT,
    -- multilinestring_col MULTILINESTRING,
    -- multipolygon_col MULTIPOLYGON,
    -- geometrycollection_col GEOMETRYCOLLECTION,

    -- JSON Data Type
    json_col JSON,

    -- Array-like representations
    set_as_array SET('value1', 'value2', 'value3', 'value4', 'value5')
);

CREATE TABLE all_types.json_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    data JSON
);
