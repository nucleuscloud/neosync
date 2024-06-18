CREATE SCHEMA IF NOT EXISTS sqlmanagerpostgres;

SET search_path TO sqlmanagerpostgres;

CREATE TABLE IF NOT EXISTS users (
    id TEXT NOT NULL,
    age int NOT NULL,
    current_salary float NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    fullname TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED
);

CREATE TABLE IF NOT EXISTS users_with_identity (
    id int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    age int NOT NULL
);


CREATE TABLE parent1 (id uuid NOT NULL DEFAULT gen_random_uuid(), CONSTRAINT pk_parent1_id PRIMARY KEY (id));
CREATE TABLE child1 (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    parent_id uuid NULL,
    CONSTRAINT pk_child1_id PRIMARY KEY (id),
    CONSTRAINT fk_child1_parent_id_parent1_id FOREIGN KEY (parent_id) REFERENCES parent1(id) ON
    DELETE
        CASCADE
);
