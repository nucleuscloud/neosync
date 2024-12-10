
CREATE TABLE IF NOT EXISTS "BadName" (
    "ID" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "NAME" text UNIQUE
);

INSERT INTO "BadName" ("NAME")
VALUES 
    ('Xk7pQ9nM3v'),
    ('Rt5wLjH2yB'),
    ('Zc8fAe4dN6'),
    ('Ym9gKu3sW5'),
    ('Vb4nPx7tJ2');

CREATE TABLE "Bad Name 123!@#" (
    "ID" SERIAL PRIMARY KEY,
    "NAME" text REFERENCES "BadName" ("NAME")
);


INSERT INTO "Bad Name 123!@#" ("NAME")
VALUES 
    ('Xk7pQ9nM3v'),
    ('Rt5wLjH2yB'),
    ('Zc8fAe4dN6'),
    ('Ym9gKu3sW5'),
    ('Vb4nPx7tJ2');
