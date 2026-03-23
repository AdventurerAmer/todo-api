CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name varchar(255) NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    is_activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1 
);