-- Google SSO: sign-in allowlist and admin flag

CREATE TABLE IF NOT EXISTS approved_users (
    email VARCHAR(255) PRIMARY KEY,
    added_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

-- Bootstrap (manual, one time):
--   INSERT INTO approved_users (email) VALUES ('tpfitz42@gmail.com');
--   -- after that account's first Google sign-in:
--   UPDATE users SET is_admin = TRUE WHERE email = 'tpfitz42@gmail.com';
