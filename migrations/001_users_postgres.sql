-- Default table name: users (configurable per project via database.user_table)
CREATE TABLE IF NOT EXISTS users (
  id                  VARCHAR(32) PRIMARY KEY,
  email               VARCHAR(255) NOT NULL UNIQUE,
  phone               VARCHAR(64) DEFAULT '',
  password_hash       TEXT NOT NULL,
  refresh_token_hash  TEXT DEFAULT '',
  reset_token_hash    TEXT DEFAULT '',
  reset_token_expiry  TIMESTAMPTZ,
  created_at          TIMESTAMPTZ NOT NULL,
  updated_at          TIMESTAMPTZ NOT NULL
);
