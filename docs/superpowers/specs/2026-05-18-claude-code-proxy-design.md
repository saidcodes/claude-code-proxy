# Claude Code Proxy — Design Spec

## Overview

A lightweight HTTP reverse proxy that lets [Claude Code CLI](https://claude.ai/code) use [DeepSeek's Anthropic-compatible API](https://api-docs.deepseek.com/guides/anthropic_api) as the backend LLM provider. No Anthropic subscription or API key required.

DeepSeek already speaks Anthropic's Messages API format, so the proxy primarily handles auth swapping, model remapping, and stripping unsupported features from the request body.

## Architecture

```
Claude Code  ──POST /v1/messages──▶  claude-code-proxy (:8080)  ──POST /v1/messages──▶  api.deepseek.com/anthropic
  ANTHROPIC_BASE_URL=http://localhost:8080                          DEEPSEEK_API_KEY=sk-...
  ANTHROPIC_API_KEY=sk-fake                                         TARGET_MODEL=deepseek-v4-flash
                                    │
                              Request transformations:
                              - Swap API key header
                              - Replace model name
                              - Strip cache_control
                              - Strip disable_parallel_tool_use
                              - Strip anthropic-beta header
```

The response (including SSE streaming) is forwarded as-is since DeepSeek returns the same Anthropic-compatible format.

## Components

**`main.go`** — Entry point. Starts HTTP server on configurable port, registers routes.

**`proxy/config.go`** — Reads configuration from environment variables into a `Config` struct with validation.

**`proxy/transforms.go`** — Pure functions that mutate the outgoing JSON request body:
- Replace `model` field with the configured DeepSeek model
- Remove `cache_control` from system prompt blocks and content blocks
- Remove `disable_parallel_tool_use` from tool_choice
- Return Anthropic-compatible error JSON shapes

**`proxy/proxy.go`** — Core reverse proxy handler:
- Receives incoming request from Claude Code
- Reads and parses the JSON body
- Applies transforms
- Builds outgoing request to DeepSeek with swapped auth header
- Forwards the response back (including streaming via `Transfer-Encoding: chunked` or SSE)
- On errors, returns JSON with Anthropic-compatible error structure

## Configuration

| Env Variable | Default | Required | Description |
|---|---|---|---|
| `DEEPSEEK_API_KEY` | — | Yes | DeepSeek API key |
| `TARGET_MODEL` | `deepseek-v4-flash` | No | Model to request from DeepSeek |
| `PROXY_PORT` | `8080` | No | Local port for the proxy |
| `TARGET_BASE_URL` | `https://api.deepseek.com/anthropic` | No | DeepSeek Anthropic endpoint |

## Request Transformations (detailed)

1. **API key swap** — Replace the `x-api-key` header value with `DEEPSEEK_API_KEY`. (Claude Code sends whatever is in `ANTHROPIC_API_KEY`, which can be a dummy value.)
2. **Model remap** — Find the `"model"` key in the JSON body and replace its value with `TARGET_MODEL`.
3. **Cache control strip** — Recursively remove any `"cache_control"` key from system items and content blocks (DeepSeek ignores it, but stripping prevents potential parsing issues).
4. **Parallel tool use strip** — Remove `"disable_parallel_tool_use"` from `"tool_choice"` if present.
5. **Header strip** — Remove `anthropic-beta` and `anthropic-version` headers (DeepSeek ignores them).

## Error Handling

- Connection errors to DeepSeek return `{"error": {"type": "api_error", "message": "..."}}` with 502 status.
- Missing `DEEPSEEK_API_KEY` causes the proxy to fail fast on startup.
- Non-`/v1/messages` and non-`/health` paths return 404.
- DeepSeek error responses are forwarded as-is since they already use Anthropic-compatible error shapes.

## Routes

| Path | Method | Purpose |
|---|---|---|
| `/health` | GET | Health check — returns 200 OK |
| `/v1/messages` | POST | Proxied Anthropic Messages API call |

## Testing

- **Unit tests** for `transforms.go` — verify model replacement, cache_control stripping, and parallel tool use stripping against sample JSON payloads.
- **Integration test** — start proxy, send a minimal `/v1/messages` request, verify response structure.

## Usage

```bash
# Start the proxy
export DEEPSEEK_API_KEY=sk-your-key-here
./claude-code-proxy &

# Point Claude Code at the proxy
export ANTHROPIC_BASE_URL=http://localhost:8080
export ANTHROPIC_API_KEY=sk-ignored

# Use Claude Code as normal
claude
```

## Non-Goals

- Not a general-purpose API gateway — only handles the `/v1/messages` endpoint.
- No support for image, document, MCP, or other content types DeepSeek doesn't support.
- No caching layer — `cache_control` is stripped, not emulated.
- No multi-model routing — all requests go to the configured `TARGET_MODEL`.
