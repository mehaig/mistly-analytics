# Mistly

A lightweight, self-hostable web analytics ingestor written in Go.

Mistly does two things:

1. Manages sites and serves a tiny JavaScript snippet (`tracker.js`) that you embed on your website.
2. Receives pageview events from that snippet and writes them to PostgreSQL.

No dashboard, no cookies, no tracking pixels. Just ingestion.

---

## Requirements

- Go 1.22+
- PostgreSQL (or a managed Postgres provider like Railway, Supabase, Neon)

---

## Deploying to Railway (recommended)

Railway is the easiest way to run Mistly — it handles HTTPS, environment variables, and the database automatically.

### 1. Fork or push this repo to GitHub

### 2. Create a Railway project

- Go to [railway.app](https://railway.app) and sign in
- Click **New Project** → **Deploy from GitHub repo** → select this repo
- Railway will start building automatically

### 3. Add a Postgres database

- Inside your Railway project, click **+ New** → **Database** → **Add PostgreSQL**
- A Postgres service appears on the canvas

### 4. Connect the database to your Go service

- Click your Go service → **Variables** tab
- Click **+ Add a variable reference** and select `DATABASE_URL` from the Postgres service
- Railway injects it automatically — you don't copy anything manually

### 5. Set the build and start commands

- Click your Go service → **Settings** tab
- Set **Custom Build Command** to:
  ```
  go build -o server ./cmd/server && go build -o mistly ./cmd/mistly
  ```
- Set **Custom Start Command** to:
  ```
  ./server
  ```
- Save — Railway will redeploy

### 6. Generate a public URL

- Still in **Settings** → scroll to **Networking** → click **Generate Domain**
- Copy the URL (e.g. `https://mistly-analytics-production.up.railway.app`)

### 7. Generate a secret admin token

Run this locally in the repo folder:

```bash
go run ./cmd/mistly token
```

Copy the token it prints.

- Go back to Railway → your Go service → **Variables** tab
- Add a new variable: `ADMIN_TOKEN` = the token you copied
- Railway redeploys automatically

### 8. Create your first site

Once the deployment shows **Active**, run this in your terminal — all on one line:

```bash
curl -X POST https://YOUR-DOMAIN.up.railway.app/sites -H "Content-Type: application/json" -H "Authorization: Bearer YOUR-TOKEN" -d '{"name": "My Site", "domain": "example.com"}'
```

It returns JSON like:

```json
{"id":"75521d6a43b97ecb","name":"My Site","domain":"example.com","created_at":"..."}
```

### 9. Get your tracker snippet

Open this URL in your browser (swap in your domain and site ID):

```
https://YOUR-DOMAIN.up.railway.app/sites/75521d6a43b97ecb/snippet
```

It returns a `<script>` tag like:

```html
<script src="https://YOUR-DOMAIN.up.railway.app/tracker.js" data-site-id="75521d6a43b97ecb" defer></script>
```

### 10. Paste the snippet into your website

Add it inside the `<head>` of every page you want to track. That's it — pageviews start flowing into your Postgres database immediately.

### Verify it's working

Open these in your browser:

- `https://YOUR-DOMAIN.up.railway.app/health` → should show `ok`
- `https://YOUR-DOMAIN.up.railway.app/tracker.js` → should show JavaScript

---

## Running locally

```bash
# Set your database URL
export DATABASE_URL="postgres://postgres:password@localhost:5432/mistly?sslmode=disable"

# Run migrations and create your first site
go run ./cmd/mistly init --name "My Site" --domain "example.com" --server-url "http://localhost:8080"

# Start the server
go run ./cmd/server
```

---

## CLI reference

Build the CLI once:

```bash
go build -o mistly ./cmd/mistly
```

### `mistly init`

Runs migrations, creates your first site, and prints the tracker snippet. Use this for local setup.

```
mistly init --name <name> [--domain <domain>] [--server-url <url>]
```

### `mistly token`

Generates a secure random token to use as `ADMIN_TOKEN`.

```
mistly token
```

### `mistly sites create`

Registers a new site and prints its snippet.

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

---

## HTTP API

All `/sites` endpoints require an `Authorization: Bearer <token>` header matching your `ADMIN_TOKEN`. The `/collect`, `/tracker.js`, and `/health` endpoints are public.

### `POST /sites`

Creates a site. Returns JSON.

```bash
curl -X POST https://YOUR-DOMAIN/sites -H "Content-Type: application/json" -H "Authorization: Bearer TOKEN" -d '{"name": "My Site", "domain": "example.com"}'
```

### `GET /sites`

Returns a JSON array of all registered sites.

```bash
curl https://YOUR-DOMAIN/sites -H "Authorization: Bearer TOKEN"
```

### `GET /sites/{id}/snippet`

Returns the tracker `<script>` tag for the given site as plain text.

```bash
curl https://YOUR-DOMAIN/sites/75521d6a43b97ecb/snippet -H "Authorization: Bearer TOKEN"
```

### `POST /collect`

Records a pageview. Called automatically by `tracker.js` — you never need to call this directly.

### `GET /tracker.js`

Serves the client-side tracker script. Embed this on your website.

### `GET /health`

Returns `ok`. Use for uptime checks.

---

## Configuration

| Variable       | Required | Default | Description                                        |
|----------------|----------|---------|----------------------------------------------------|
| `DATABASE_URL` | yes      | —       | Postgres connection string                         |
| `ADMIN_TOKEN`  | yes      | —       | Secret token required to access `/sites` endpoints |
| `PORT`         | no       | `8080`  | Port to listen on                                  |

---

## What data is collected

Every pageview record contains:

| Field | Source |
|---|---|
| `url` | Sent by tracker |
| `referrer` | Sent by tracker |
| `page_title` | Sent by tracker (`document.title`) |
| `screen_width` / `screen_height` | Sent by tracker (`window.screen`) |
| `browser` / `os` / `device` | Derived server-side from `User-Agent` |
| `language` | Derived server-side from `Accept-Language` header |
| `utm_source` / `utm_medium` / `utm_campaign` | Parsed server-side from the page URL |
| `session_id` | `sha256(ip + user_agent + site_id + date)` — resets daily, no cookies |

---

## Self-hosting on a VPS

Put Mistly behind a reverse proxy (nginx, Caddy) that forwards `X-Forwarded-For` so session hashing uses the real client IP. On managed platforms like Railway, Render, or Fly.io this is handled automatically.

`tracker.js` is compiled into the server binary — no static files needed alongside it.

---

## License

MIT
