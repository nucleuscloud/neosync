CREATE SCHEMA IF NOT EXISTS javascript;
SET search_path TO javascript;

CREATE TABLE transformers (
    id SERIAL PRIMARY KEY,
    e164_phone_number VARCHAR(20),
    email VARCHAR(255),
    measurement FLOAT,
    int64 BIGINT,
    int64_phone_number BIGINT,
    string_phone_number VARCHAR(20),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    full_name VARCHAR(255),
    str VARCHAR (255),
    character_scramble VARCHAR (255),
    bool BOOLEAN,
    card_number BIGINT,
    categorical VARCHAR(255),
    city VARCHAR(255),
    full_address VARCHAR(255),
    gender VARCHAR(255),
    international_phone VARCHAR(255),
    sha256 VARCHAR(255),
    ssn VARCHAR(255),
    state VARCHAR(255),
    street_address VARCHAR(255),
    unix_time BIGINT,
    username VARCHAR(255),
    utc_timestamp TIMESTAMPTZ,
    uuid VARCHAR(255),
    zipcode BIGINT
);
