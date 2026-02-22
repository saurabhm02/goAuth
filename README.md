# goAuth

JWT auth service in Go with multi-project support; for now PostgreSQL only. One server, many projects; each project has its own DB. Send **`X-Project-ID`** (default: `default`) on every request.

## Setup

**Env** (copy `.env.example` → `.env`):

- **PORT** (default `8080`) · **JWT_SECRET** · **DATABASE_DSN** (Postgres URL for project `default`) · **CONFIG_PATH** (optional)
- When any project has **otp: true**: **SMTP_HOST**, **SMTP_PORT**, **SMTP_USER**, **SMTP_PASSWORD**, **SMTP_FROM**

**config.yaml** — projects, optional DSN override, and optional OTP:

```yaml
projects:
  default:
    database:
      dsn: ""
      user_table: users
      otp_table: otp  
    otp: false         # set true to enable OTP (email); requires SMTP_* env
```

When `otp: true`: signup sends OTP to email; login can use password or `use_otp: true` (then verify-otp to get token). Set **SMTP_HOST**, **SMTP_PORT**, **SMTP_USER**, **SMTP_PASSWORD**, **SMTP_FROM** in env.

## Database

**users** (required):

```sql
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
```

**otp** (only when project has `otp: true`):

```sql
CREATE TABLE IF NOT EXISTS otp (
  id        VARCHAR(32) PRIMARY KEY,
  user_id   VARCHAR(32) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  otp       TEXT NOT NULL,
  expiry    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_otp_user_id ON otp(user_id);
CREATE INDEX IF NOT EXISTS idx_otp_expiry ON otp(expiry);
```

## Run

```bash
export $(grep -v '^#' .env | xargs)
go run ./cmd/server
```

## API

| Method | Path | Body | Description |
|--------|------|------|-------------|
| POST | /auth/signup | `{"email","phone","password"}` | Register (when OTP: returns `message: "otp_sent"`) |
| POST | /auth/login | `{"email_or_phone","password"}` or `{"email_or_phone","use_otp":true}` | Login → `user`, `token`, `expires_in` or `message: "otp_sent"` |
| POST | /auth/verify-otp | `{"email_or_phone","otp"}` | Verify OTP → `user`, `token`, `expires_in` (when project has OTP) |
| GET | /auth/me | — | Current user (`Authorization: Bearer <token>`) |
| GET | /health | — | Health check |

Success: `{"data": ...}` · Error: `{"error": "message"}`
