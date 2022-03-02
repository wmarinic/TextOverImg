BEGIN;

CREATE SCHEMA IF NOT EXISTS inspirationifier;

CREATE TABLE inspirationifier.users (
    User_ID         BIGSERIAL PRIMARY KEY,
    User_Name       TEXT NOT NULL,
    Password_Hash   TEXT NOT NULL,
    UNIQUE (User_Name)
);

COMMIT;