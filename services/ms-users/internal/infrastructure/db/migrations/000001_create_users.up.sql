CREATE TABLE users (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Unique email only among active rows (soft delete allows reusing email).
CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL;
