CREATE TABLE IF NOT EXISTS payments (
    id            VARCHAR(64) PRIMARY KEY,
    amount        BIGINT      NOT NULL,
    currency      VARCHAR(3)  NOT NULL DEFAULT 'THB',
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid          BOOLEAN     NOT NULL DEFAULT FALSE,
    description   TEXT,
    idempotency_key VARCHAR(128) UNIQUE,
    omise_charge_id VARCHAR(64),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
