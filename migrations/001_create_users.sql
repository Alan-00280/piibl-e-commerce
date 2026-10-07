-- Tabel users
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL       PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL,
    email      VARCHAR(255) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE TYPE user_role AS ENUM (
    'ADMIN',
    'CUSTOMER',
    'TENANT'
);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role user_role NOT NULL;
 
-- Keunikan username tanpa membedakan huruf besar dan kecil.
-- Inilah yang menggantikan pemeriksaan manual di pertemuan 2
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_key
    ON users (LOWER(username));
 
CREATE INDEX IF NOT EXISTS users_email_lower_idx
    ON users (LOWER(email));
