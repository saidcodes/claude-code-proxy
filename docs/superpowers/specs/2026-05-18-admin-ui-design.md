# Admin UI — Design Spec

## Overview

Add a web-based admin interface to the proxy so users can input their DeepSeek API key through a browser instead of setting environment variables. The key is persisted to disk and survives restarts.

## Routes

| Route | Method | Purpose |
|---|---|---|
| `GET /` | GET | Admin page — key form or status |
| `POST /admin/key` | POST | Save API key to disk |
| `GET /health` | GET | Health check (unchanged) |
| `POST /v1/messages` | POST | Messages API (unchanged) |

## Startup Flow

1. Load `DEEPSEEK_API_KEY` from env var — if set, proxy is ready immediately
2. If not set, read `~/.claude-code-proxy/key.json` — if file exists, load the key
3. If neither source provides a key, start in "setup needed" mode — the admin page is active, and `/v1/messages` returns an error asking the user to set a key via the admin page

## Key File

- **Path:** `~/.claude-code-proxy/key.json` (platform-appropriate home directory)
- **Format:** `{"api_key":"sk-..."}`
- **Persistence:** Read on startup, written on POST /admin/key

## Admin HTML Page

A single HTML page embedded into the Go binary via `embed.FS`:
- If no key is set: shows a form with a password-style API key input and "Save Key" button
- If key is set: shows a status page indicating the proxy is ready, with the key masked
- Uses `fetch()` to POST to `/admin/key` — no page reload, shows success/error inline

## Code Changes

**New file: `proxy/admin.go`** — Admin HTTP handler:
- `AdminHandler` struct with embedded HTML
- GET `/` serves the admin page (key form or status page)
- POST `/admin/key` parses `{"api_key":"..."}`, calls `SaveKeyFile()`, updates Config

**New file: `proxy/admin.html`** — HTML page embedded via `//go:embed`

**New file: `proxy/keyfile.go`** — Key file read/write functions:
- `LoadKeyFile() (string, error)` — reads key from `~/.claude-code-proxy/key.json`
- `SaveKeyFile(key string) error` — writes key to `~/.claude-code-proxy/key.json`
- `keyFilePath() string` — returns the platform key file path

**Modified: `proxy/config.go`** — Add:
- `SetAPIKey(key string)` method on Config (sets field in memory)

**Modified: `proxy/proxy.go`** — Modify:
- `ServeHTTP` checks if `DeepSeekAPIKey` is empty before forwarding
- If empty, returns error asking user to visit the admin page
- Register admin routes in a new `RegisterAdminRoutes` or similar

**Modified: `main.go`** — Modify startup:
- Load env config (no longer fatal if DEEPSEEK_API_KEY is missing)
- Try key file as fallback
- Start server regardless

## Testing

- `keyfile_test.go` — Unit tests for LoadKeyFile/SaveKeyFile with temp dirs
- `admin_test.go` — Tests for GET/ and POST /admin/key with httptest
- Integration check: proxy starts without env var and serves admin page
