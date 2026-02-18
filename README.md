# goAuth

JWT auth service in Go (PostgreSQL). Multi-tenant: one server, many projects; each project has its own DB. Send **`X-Project-ID`** (default: `default`) on every request.

## Setup

**Env** (copy `.env.example` → `.env`):

- **PORT** (default `8080`) · **JWT_SECRET** · **DATABASE_DSN** (Postgres URL for project `default`) · **CONFIG_PATH** (optional)

**config.yaml** — projects and optional DSN override (else uses `DATABASE_DSN`):

```yaml
projects:
  default:
    database:
      dsn: ""
      user_table: users
```

## Database

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

## Run

```bash
export $(grep -v '^#' .env | xargs)
go run ./cmd/server
```

## API

| Method | Path | Body | Description |
|--------|------|------|-------------|
| POST | /auth/signup | `{"email","phone","password"}` | Register |
| POST | /auth/login | `{"email_or_phone","password"}` | Login → `user`, `token`, `expires_in` |
| GET | /auth/me | — | Current user (`Authorization: Bearer <token>`) |
| GET | /health | — | Health check |

Success: `{"data": ...}` · Error: `{"error": "message"}`
