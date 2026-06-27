-- Table of users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK ( role IN ('admin', 'user') ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);