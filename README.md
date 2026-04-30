# Mistly

A lightweight, self-hostable web analytics ingestor written in Go.

Mistly does two things:

1. Manages sites and serves a tiny JavaScript snippet (`tracker.js`) that you embed on your website.
2. Receives pageview events from that snippet and writes them to PostgreSQL.

No dashboard, no cookies, no tracking pixels. Just ingestion.

## Requirements

- Go 1.22+
- PostgreSQL

## Quick start

```bash
# 1. Set your database connection string
export DATABASE_URL="postgres://postgres:password@localhost:5432/mistly?sslmode=disable"

# 2. Initialise the database, create your first site, and get your snippet
go run ./cmd/mistly init \
  --name "My Site" \
  --domain "example.com" \
  --server-url "https://analytics.example.com"

# 3. Start the server (must be run from the repo root)
go run ./cmd/server
```

`init` creates the database tables, registers your site, and prints the `<script>` tag ready to paste into your website's `<head>`.

## CLI

The `mistly` binary is the management tool. Build it once:

```bash
go build -o mistly ./cmd/mistly
```

### `mistly init`

Connects to the database, runs migrations, creates your first site, and prints its snippet.

```
mistly init --name <name> [--domain <domain>] [--server-url <url>]

Flags:
  --name        Site name (required)
  --domain      Site domain, e.g. example.com (optional)
  --server-url  Public URL of your Mistly server (default: http://localhost:8080)
```

### `mistly sites create`

Adds another site and prints its snippet.

```
mistly sites create --name <name> [--domain <domain>] [--server-url <url>]
```

### `mistly sites list`

Lists all registered sites.

```
mistly sites list
```

### `mistly sites snippet`

Prints the tracker snippet for an existing site.

```
mistly sites snippet <id> [--server-url <url>]
```

## HTTP API

### `POST /sites`

Creates a site. Returns JSON.

```bash
curl -X POST http://localhost:8080/sites \
  -H "Content-Type: application/json" \
  -d '{"name": "My Site", "domain": "example.com"}'
```

### `GET /sites`

Returns a JSON array of all registered sites.

### `GET /sites/{id}/snippet`

Returns the tracker `<script>` tag for the given site as plain text.

### `POST /collect`

Records a pageview. Called automatically by `tracker.js` — you do not need to call this directly.

```json
{
  "site_id": "a3f8c2d1e4b56789",
  "url": "https://example.com/page",
  "referrer": "https://google.com",
  "page_title": "My Page",
  "screen_width": 1440,
  "screen_height": 900
}
```

- `site_id` and `url` are required.
- `browser`, `os`, `device`, `language`, and UTM parameters are derived server-side.
- Session ID is derived from `sha256(ip | user_agent | site_id | YYYY-MM-DD)` — no cookies, no persistent identifier.
- Returns `204 No Content` on success.

### `GET /tracker.js`

Serves the client-side snippet with `Cache-Control: public, max-age=3600`.

### `GET /health`

Returns `ok`. Use this for uptime checks.

## Configuration

| Variable       | Required | Default | Description                |
|----------------|----------|---------|----------------------------|
| `DATABASE_URL` | yes      | —       | Postgres connection string |
| `PORT`         | no       | `8080`  | Port to listen on          |

## Schema

```sql
CREATE TABLE sites (
    id         TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    domain     TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE pageviews (
    id            BIGSERIAL   PRIMARY KEY,
    site_id       TEXT        NOT NULL,
    url           TEXT        NOT NULL,
    referrer      TEXT,
    browser       TEXT,
    os            TEXT,
    device        TEXT,
    session_id    TEXT,
    language      TEXT,
    page_title    TEXT,
    screen_width  INT,
    screen_height INT,
    utm_source    TEXT,
    utm_medium    TEXT,
    utm_campaign  TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
```

Indexes exist on `pageviews(site_id)` and `pageviews(created_at)`.

Migrations are idempotent and run automatically on every server or CLI start — no migration tool needed.

## Deployment

`tracker.js` is compiled directly into the server binary via `//go:embed` — you can run the binary from any directory, no `static/` folder needed alongside it.

For the most straightforward deployment, use a platform like **Railway**:

1. Connect your repo — Railway auto-detects Go and builds it
2. Add a Postgres plugin — `DATABASE_URL` is injected automatically
3. Deploy — Railway provides an HTTPS domain with the proxy already configured correctly
4. Open the Railway shell and run `mistly init` once to create your tables and first site

If you self-host on a VPS, put Mistly behind a reverse proxy (nginx, Caddy) that forwards `X-Forwarded-For` so the session hash reflects the real client IP. On managed platforms (Railway, Render, Fly.io) this is handled for you.

## License

MIT
