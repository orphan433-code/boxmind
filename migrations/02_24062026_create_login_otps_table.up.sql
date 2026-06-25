CREATE TABLE IF NOT EXISTS login_otps (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT        NOT NULL,
    code_hash  TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_login_otps_email ON login_otps (email);
CREATE INDEX IF NOT EXISTS idx_login_otps_expires_at ON login_otps (expires_at);
