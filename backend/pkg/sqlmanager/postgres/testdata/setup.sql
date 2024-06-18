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

-- testing basic circular deps
CREATE TABLE t1 (
    a int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    b int NULL,
    CONSTRAINT t1_b_fkey FOREIGN KEY (b) REFERENCES t1(a)
);
CREATE TABLE t2 (
    a int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    b int NULL
);
CREATE TABLE t3 (
    a int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    b int NULL
);
ALTER TABLE t2
ADD CONSTRAINT t2_b_fkey FOREIGN KEY (b) REFERENCES t3(a);
ALTER TABLE t3
ADD CONSTRAINT t3_b_fkey FOREIGN KEY (b) REFERENCES t2(a);

-- Testing composite keys
CREATE TABLE t4 (
    a int NOT NULL,
    b int NOT NULL,
    c int NULL,
    PRIMARY KEY (a, b)
);
CREATE TABLE t5 (
    x int NOT NULL,
    y int NOT NULL,
    z int NULL,
    CONSTRAINT t5_t4_fkey FOREIGN KEY (x, y) REFERENCES t4 (a, b)
);
