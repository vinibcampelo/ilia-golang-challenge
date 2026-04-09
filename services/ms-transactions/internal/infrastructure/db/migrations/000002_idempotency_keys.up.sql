CREATE TABLE idempotency_keys (
    user_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    request_fingerprint TEXT NOT NULL,
    transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, idempotency_key)
);
