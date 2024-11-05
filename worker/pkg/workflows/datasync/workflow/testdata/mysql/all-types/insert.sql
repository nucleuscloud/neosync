USE all_types;

INSERT INTO all_data_types (
    tinyint_col, smallint_col, mediumint_col, int_col, bigint_col,
    decimal_col, float_col, double_col, 
    bit_col,
    date_col,
    time_col, datetime_col, timestamp_col, year_col,
    char_col, varchar_col,binary_col, varbinary_col,
    tinyblob_col, tinytext_col, blob_col, text_col,
    mediumblob_col, mediumtext_col, longblob_col, longtext_col,
    enum_col, set_col,
    -- geometry_col, point_col, linestring_col, polygon_col,
    -- multipoint_col, multilinestring_col, multipolygon_col, geometrycollection_col,
    json_col,
    set_as_array
) VALUES (
    127, 32767, 8388607, 2147483647, 9223372036854775807,
    1234.56, 3.1415, 3.14159265359, 
    b'10101010',
    '2023-09-12', '14:30:00', '2023-09-12 14:30:00', 2023,
    'Fixed Char', 'Variable Char', 'Bin', 'VarBinary',
    'Tiny BLOB', 'Tiny Text', 'Regular BLOB', 'Regular Text',
    'Medium BLOB', 'Medium Text', 'Long BLOB', 'Long Text',
    'value2', 'option1,option3',
    -- ST_GeomFromText('POINT(1 1)'),
    -- ST_PointFromText('POINT(1 1)'),
    -- ST_LineFromText('LINESTRING(0 0,1 1,2 2)'),
    -- ST_PolygonFromText('POLYGON((0 0,10 0,10 10,0 10,0 0),(5 5,7 5,7 7,5 7,5 5))'),
    -- ST_MultiPointFromText('MULTIPOINT(1 1, 2 2)'),
    -- ST_MultiLineStringFromText('MULTILINESTRING((0 0,1 1,2 2),(2 2,3 3,4 4))'),
    -- ST_MultiPolygonFromText('MULTIPOLYGON(((0 0,10 0,10 10,0 10,0 0)),((5 5,7 5,7 7,5 7,5 5)))'),
    -- ST_GeomCollFromText('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1,2 2))'),
    '{"key": "value", "array": [1, 2, 3]}',
    'value1,value3,value5'
);

INSERT INTO all_data_types (id) VALUES (DEFAULT);


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
