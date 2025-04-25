CREATE SCHEMA IF NOT EXISTS auth_schema;
CREATE TABLE IF NOT EXISTS auth_schema.users
(
    id SERIAL PRIMARY KEY,
    token TEXT,
    login VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(30) UNIQUE,
    pass TEXT NOT NULL,
    name VARCHAR(100)
);