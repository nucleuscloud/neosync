CREATE TABLE IF NOT EXISTS all_data_types (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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
    -- time_col TIME,
    -- timetz_col TIMETZ,
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


INSERT INTO all_data_types (
    Id,
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
    -- time_col, 
    -- timetz_col, 
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
    DEFAULT,
    32767,  -- smallint_col
    2147483647,  -- integer_col
    922337203685477580,  -- bigint_col
    1234.56,  -- decimal_col
    99999999.99,  -- numeric_col
    12345.67,  -- real_col
    123456789.123456789,  -- double_precision_col
    1,  -- serial_col (auto-incremented, will be generated)
    1,  -- bigserial_col (auto-incremented, will be generated)
    '$100.00',  -- money_col
    'A',  -- char_col
    'DEFAULT',  -- varchar_col
    'default',  -- text_col
    decode('DEADBEEF', 'hex'),  -- bytea_col
    '2024-01-01 12:34:56',  -- timestamp_col
    '2024-01-01 12:34:56+00',  -- timestamptz_col
    '2024-01-01',  -- date_col
    -- '12:34:56',  -- time_col
    -- '12:34:56+00',  -- timetz_col
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

INSERT INTO all_data_types (
    Id
) VALUES (
    DEFAULT
);


-- CREATE TABLE IF NOT EXISTS time_time (
--     id SERIAL PRIMARY KEY,
--     timestamp_col TIMESTAMP,
--     timestamptz_col TIMESTAMPTZ,
--     date_col DATE
-- );

-- INSERT INTO time_time (
--     timestamp_col,
--     timestamptz_col,
--     date_col
-- ) 
-- VALUES (
--     '2024-03-18 10:30:00',
--     '2024-03-18 10:30:00+00',
--     '2024-03-18'
-- );

-- INSERT INTO time_time (
--     timestamp_col,
--     timestamptz_col,
--     date_col
-- ) 
-- VALUES (
--     '0001-01-01 00:00:00 BC',
--     '0001-01-01 00:00:00+00 BC',
--     '0001-01-01 BC'
-- );


CREATE TABLE IF NOT EXISTS array_types (
    "id" BIGINT NOT NULL PRIMARY KEY,
    -- "int_array" _int4,
    -- "smallint_array" _int2,
    -- "bigint_array" _int8,
    -- "real_array" _float4,
    -- "double_array" _float8,
    -- "text_array" _text,
    -- "varchar_array" _varchar,
    -- "char_array" _bpchar,
    -- "boolean_array" _bool,
    -- "date_array" _date,
    -- "time_array" _time,
    -- "timestamp_array" _timestamp,
    -- "timestamptz_array" _timestamptz,
    "interval_array" _interval
    -- "inet_array" _inet, // broken
    -- "cidr_array" _cidr,
    -- "point_array" _point,
    -- "line_array" _line,
    -- "lseg_array" _lseg,
    -- "box_array" _box,   // broken
    -- "path_array" _path,
    -- "polygon_array" _polygon,
    -- "circle_array" _circle,
    -- "uuid_array" _uuid,
    -- "json_array" _json,
    -- "jsonb_array" _jsonb,
    -- "bit_array" _bit,
    -- "varbit_array" _varbit,
    -- "numeric_array" _numeric,
    -- "money_array" _money,
    -- "xml_array" _xml
    -- "int_double_array" _int4
);


INSERT INTO array_types (
    id, 
    -- int_array, smallint_array, bigint_array,
    -- real_array,
    -- double_array,
    -- text_array, varchar_array, char_array, boolean_array,
    -- date_array,
    -- time_array, timestamp_array, timestamptz_array,
    interval_array
    -- inet_array, cidr_array, 
    -- point_array, line_array, lseg_array,
    -- box_array,
    -- path_array, polygon_array, circle_array, 
    -- uuid_array, 
    -- json_array, jsonb_array, 
    -- bit_array, varbit_array, numeric_array,
    -- money_array,
    -- xml_array
    -- int_double_array
) VALUES (
    1,
    -- ARRAY[1, 2, 3],
    -- ARRAY[10::smallint, 20::smallint],
    -- ARRAY[100::bigint, 200::bigint],
    -- ARRAY[1.1::real, 2.2::real],
    -- ARRAY[1.11::double precision, 2.22::double precision],
    -- ARRAY['text1', 'text2'],
    -- ARRAY['varchar1'::varchar, 'varchar2'::varchar],
    -- ARRAY['a'::char, 'b'::char],
    -- ARRAY[true, false],
    -- ARRAY['2023-01-01'::date, '2023-01-02'::date],
    -- ARRAY['12:00:00'::time, '13:00:00'::time],
    -- ARRAY['2023-01-01 12:00:00'::timestamp, '2023-01-02 13:00:00'::timestamp],
    -- ARRAY['2023-01-01 12:00:00+00'::timestamptz, '2023-01-02 13:00:00+00'::timestamptz],
    ARRAY['1 day'::interval, '2 hours'::interval]
    -- ARRAY['192.168.0.1'::inet, '10.0.0.1'::inet],
    -- ARRAY['192.168.0.0/24'::cidr, '10.0.0.0/8'::cidr],
    -- ARRAY['(1,1)'::point, '(2,2)'::point],
    -- ARRAY['{1,2,2}'::line, '{3,4,4}'::line],
    -- ARRAY['(1,1,2,2)'::lseg, '(3,3,4,4)'::lseg],
    -- ARRAY['(1,1,2,2)'::box, '(3,3,4,4)'::box],
    -- ARRAY['((1,1),(2,2),(3,3))'::path, '((4,4),(5,5),(6,6))'::path],
    -- ARRAY['((1,1),(2,2),(3,3))'::polygon, '((4,4),(5,5),(6,6))'::polygon],
    -- ARRAY['<(1,1),1>'::circle, '<(2,2),2>'::circle],
    -- ARRAY['a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid, 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid],
    -- ARRAY['{"key": "value1"}'::json, '{"key": "value2"}'::json],
    -- ARRAY['{"key": "value1"}'::jsonb, '{"key": "value2"}'::jsonb],
    -- ARRAY['101'::bit(3), '110'::bit(3)],
    -- ARRAY['10101'::bit varying(5), '01010'::bit varying(5)],
    -- ARRAY[1.23::numeric, 4.56::numeric],
    -- ARRAY[10.00::money, 20.00::money],
    -- ARRAY['<root>value1</root>'::xml, '<root>value2</root>'::xml]
    -- ARRAY[[1, 2], [3, 4]] 
);


CREATE TABLE json_data (
    id SERIAL PRIMARY KEY,
    data JSONB
);


INSERT INTO json_data (data) VALUES ('"Hello, world!"');
INSERT INTO json_data (data) VALUES ('42');
INSERT INTO json_data (data) VALUES ('3.14');
INSERT INTO json_data (data) VALUES ('true');
INSERT INTO json_data (data) VALUES ('false');
INSERT INTO json_data (data) VALUES ('null');

INSERT INTO json_data (data) VALUES ('{"name": "John", "age": 30}');
INSERT INTO json_data (data) VALUES ('{"coords": {"x": 10, "y": 20}}');

INSERT INTO json_data (data) VALUES ('[1, 2, 3, 4]');
INSERT INTO json_data (data) VALUES ('["apple", "banana", "cherry"]');

INSERT INTO json_data (data) VALUES ('{"items": ["book", "pen"], "count": 2, "in_stock": true}');

INSERT INTO json_data (data) VALUES (
    '{
        "user": {
            "name": "Alice",
            "age": 28,
            "contacts": [
                {"type": "email", "value": "alice@example.com"},
                {"type": "phone", "value": "123-456-7890"}
            ]
        },
        "orders": [
            {"id": 1001, "total": 59.99},
            {"id": 1002, "total": 24.50}
        ],
        "preferences": {
            "notifications": true,
            "theme": "dark"
        }
    }'
);
