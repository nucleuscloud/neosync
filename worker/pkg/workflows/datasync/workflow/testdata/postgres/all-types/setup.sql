CREATE SCHEMA IF NOT EXISTS ptypes;
CREATE TABLE ptypes.all_postgres_types (
    -- Numeric Types
    smallint_col SMALLINT,
    integer_col INTEGER,
    bigint_col BIGINT,
    decimal_col DECIMAL(10, 2),
    numeric_col NUMERIC(10, 2),
    real_col REAL,
    double_precision_col DOUBLE PRECISION,
    serial_col SERIAL,
    bigserial_col BIGSERIAL,

    -- Monetary Types
    money_col MONEY,

    -- Character Types
    char_col CHAR(10),
    varchar_col VARCHAR(50),
    text_col TEXT,

    -- Binary Types
    bytea_col BYTEA,

    -- Date/Time Types
    timestamp_col TIMESTAMP,
    timestamptz_col TIMESTAMPTZ,
    date_col DATE,
    time_col TIME,
    timetz_col TIMETZ,
    interval_col INTERVAL,

    -- Boolean Type
    boolean_col BOOLEAN,

    -- UUID Type
    uuid_col UUID,

    -- Network Address Types
    inet_col INET,
    cidr_col CIDR,
    macaddr_col MACADDR,

    -- Bit String Types
    bit_col BIT(8),
    varbit_col VARBIT(8),

    -- Geometric Types
    point_col POINT,
    line_col LINE,
    lseg_col LSEG,
    box_col BOX,
    path_col PATH,
    polygon_col POLYGON,
    circle_col CIRCLE,

    -- JSON Types
    json_col JSON,
    jsonb_col JSONB,

    -- Range Types
    int4range_col INT4RANGE,
    int8range_col INT8RANGE,
    numrange_col NUMRANGE,
    tsrange_col TSRANGE,
    tstzrange_col TSTZRANGE,
    daterange_col DATERANGE,

    -- Array Types
    integer_array_col INTEGER[],
    text_array_col TEXT[],

    -- XML Type
    xml_col XML,

    -- TSVECTOR Type (Full-Text Search)
    tsvector_col TSVECTOR,

    -- OID Type
    oid_col OID
);


INSERT INTO ptypes.all_postgres_types (
    smallint_col, 
    integer_col, 
    bigint_col, 
    decimal_col, 
    numeric_col, 
    real_col, 
    double_precision_col, 
    serial_col, 
    bigserial_col, 
    money_col, 
    char_col, 
    varchar_col, 
    text_col, 
    bytea_col, 
    timestamp_col, 
    timestamptz_col, 
    date_col, 
    time_col, 
    timetz_col, 
    interval_col, 
    boolean_col, 
    uuid_col, 
    inet_col, 
    cidr_col, 
    macaddr_col, 
    bit_col, 
    varbit_col, 
    point_col, 
    line_col, 
    lseg_col, 
    box_col, 
    path_col, 
    polygon_col, 
    circle_col, 
    json_col, 
    jsonb_col, 
    int4range_col, 
    int8range_col, 
    numrange_col, 
    tsrange_col, 
    tstzrange_col, 
    daterange_col, 
    integer_array_col, 
    text_array_col, 
    xml_col, 
    tsvector_col, 
    oid_col
) VALUES (
    32767,  -- smallint_col
    2147483647,  -- integer_col
    9223372036854775807,  -- bigint_col
    1234.56,  -- decimal_col
    99999999.99,  -- numeric_col
    12345.67,  -- real_col
    123456789.123456789,  -- double_precision_col
    1,  -- serial_col (auto-incremented, will be generated)
    1,  -- bigserial_col (auto-incremented, will be generated)
    '$100.00',  -- money_col
    'A',  -- char_col
    'Example varchar',  -- varchar_col
    'Example text',  -- text_col
    decode('DEADBEEF', 'hex'),  -- bytea_col
    '2024-01-01 12:34:56',  -- timestamp_col
    '2024-01-01 12:34:56+00',  -- timestamptz_col
    '2024-01-01',  -- date_col
    '12:34:56',  -- time_col
    '12:34:56+00',  -- timetz_col
    '1 day',  -- interval_col
    TRUE,  -- boolean_col
    '123e4567-e89b-12d3-a456-426614174000',  -- uuid_col
    '192.168.1.1',  -- inet_col
    '192.168.1.0/24',  -- cidr_col
    '08:00:2b:01:02:03',  -- macaddr_col
    B'10101010',  -- bit_col
    B'1010',  -- varbit_col
    '(1, 2)',  -- point_col
    '{1, 1, 0}',  -- line_col
    '[(0,0), (1,1)]',  -- lseg_col
    '(0,0),(1,1)',  -- box_col
    '((0,0), (1,1), (2,2))',  -- path_col
    '((0,0), (1,1), (1,0))',  -- polygon_col
    '<(1,1),1>',  -- circle_col
    '{"name": "John", "age": 30}',  -- json_col
    '{"name": "John", "age": 30}',  -- jsonb_col
    '[1,10]',  -- int4range_col
    '[1,1000]',  -- int8range_col
    '[1.0,10.0]',  -- numrange_col
    '[2024-01-01 12:00:00, 2024-01-01 13:00:00]',  -- tsrange_col
    '[2024-01-01 12:00:00+00, 2024-01-01 13:00:00+00]',  -- tstzrange_col
    '[2024-01-01, 2024-01-02]',  -- daterange_col
    '{1, 2, 3}',  -- integer_array_col
    '{"one", "two", "three"}',  -- text_array_col
    '<foo>bar</foo>',  -- xml_col
    'example tsvector',  -- tsvector_col
    123456  -- oid_col
);


INSERT INTO ptypes.all_postgres_types (
    smallint_col, 
    integer_col, 
    bigint_col, 
    decimal_col, 
    numeric_col, 
    real_col, 
    double_precision_col, 
    serial_col, 
    bigserial_col, 
    money_col, 
    char_col, 
    varchar_col, 
    text_col, 
    bytea_col, 
    timestamp_col, 
    timestamptz_col, 
    date_col, 
    time_col, 
    timetz_col, 
    interval_col, 
    boolean_col, 
    uuid_col, 
    inet_col, 
    cidr_col, 
    macaddr_col, 
    bit_col, 
    varbit_col, 
    point_col, 
    line_col, 
    lseg_col, 
    box_col, 
    path_col, 
    polygon_col, 
    circle_col, 
    json_col, 
    jsonb_col, 
    int4range_col, 
    int8range_col, 
    numrange_col, 
    tsrange_col, 
    tstzrange_col, 
    daterange_col, 
    integer_array_col, 
    text_array_col, 
    xml_col, 
    tsvector_col, 
    oid_col
) VALUES (
    -32768,  -- smallint_col
    -2147483648,  -- integer_col
    -9223372036854775808,  -- bigint_col
    5678.90,  -- decimal_col
    12345678.90,  -- numeric_col
    54321.12,  -- real_col
    987654321.987654321,  -- double_precision_col
    2,  -- serial_col
    2,  -- bigserial_col
    '$200.00',  -- money_col
    'B',  -- char_col
    'Another varchar example',  -- varchar_col
    'Another text example',  -- text_col
    decode('BAADF00D', 'hex'),  -- bytea_col
    '2023-12-31 23:59:59',  -- timestamp_col
    '2023-12-31 23:59:59+00',  -- timestamptz_col
    '2023-12-31',  -- date_col
    '23:59:59',  -- time_col
    '23:59:59+00',  -- timetz_col
    '2 hours',  -- interval_col
    FALSE,  -- boolean_col
    '223e4567-e89b-12d3-a456-426614174001',  -- uuid_col
    '10.0.0.1',  -- inet_col
    '10.0.0.0/24',  -- cidr_col
    '08:00:2b:01:02:04',  -- macaddr_col
    B'11001100',  -- bit_col
    B'1100',  -- varbit_col
    '(2, 3)',  -- point_col
    '{2, 2, 1}',  -- line_col
    '[(1,1), (2,2)]',  -- lseg_col
    '(1,1),(2,2)',  -- box_col
    '((1,1), (2,2), (3,3))',  -- path_col
    '((1,1), (2,2), (2,1))',  -- polygon_col
    '<(2,2),2>',  -- circle_col
    '{"name": "Jane", "age": 25}',  -- json_col
    '{"name": "Jane", "age": 25}',  -- jsonb_col
    '[2,20]',  -- int4range_col
    '[2,2000]',  -- int8range_col
    '[2.0,20.0]',  -- numrange_col
    '[2023-12-31 12:00:00, 2023-12-31 14:00:00]',  -- tsrange_col
    '[2023-12-31 12:00:00+00, 2023-12-31 14:00:00+00]',  -- tstzrange_col
    '[2023-12-31, 2024-01-01]',  -- daterange_col
    '{4, 5, 6}',  -- integer_array_col
    '{"four", "five", "six"}',  -- text_array_col
    '<baz>qux</baz>',  -- xml_col
    'another tsvector example',  -- tsvector_col
    654321  -- oid_col
);
