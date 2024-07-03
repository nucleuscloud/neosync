CREATE SCHEMA IF NOT EXISTS javascript;
SET search_path TO javascript;

CREATE TABLE transformers (
    id SERIAL PRIMARY KEY,
    e164_phone_number VARCHAR(15),
    email VARCHAR(255),
    measurement FLOAT,
    int64 BIGINT,
    int64_phone_number BIGINT,
    string_phone_number VARCHAR(15),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    full_name VARCHAR(255),
    str VARCHAR (255),
    character_scramble VARCHAR (255)
);
