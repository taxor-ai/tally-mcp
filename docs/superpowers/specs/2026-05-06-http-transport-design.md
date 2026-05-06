# Design: HTTP Transport for tally-mcp

**Date:** 2026-05-06  
**Status:** Approved  
**Requirement:** `docs/requirements/http-transport.md`

---

## Context

tally-mcp currently communicates via stdio (JSON-RPC 2.0 over stdin/stdout), which works with Claude Desktop but not with web-based MCP clients. Open WebUI (v0.6.31+) supports only Streamable HTTP transport for MCP servers. This design adds native HTTP transport so tally-mcp works with Open WebUI without requiring the mcpo proxy.

Existing stdio transport must continue to work unchanged.

---

## Transport Selection

Transport is selected by the presence of `MCP_HTTP_PORT`:

- **Unset** → stdio mode (existing behaviour, backward compatible)
- **Set** → HTTP mode, binds to `MCP_HTTP_HOST:MCP_HTTP_PORT`

This is checked once at startup in `main()`. No runtime switching.

---

## Configuration Changes (`pkg/config/config.go`)

Add `HTTPConfig` to the existing `Config` struct:

```go
type HTTPConfig struct {
    Port string // empty string = stdio mode; e.g. "9090"
    Host string // bind address, default "0.0.0.0"
}

type Config struct {
    Tally  TallyConfig
    Logger LoggerConfig
    HTTP   HTTPConfig   // new
}
```

New environment variables loaded in `config.Load()`:

| Variable       | Default   | Description                              |
|----------------|-----------|------------------------------------------|
| `MCP_HTTP_PORT`| (unset)   | Port to listen on. Unset = stdio mode.   |
| `MCP_HTTP_HOST`| `0.0.0.0` | Interface to bind to.                    |

No validation errors if `MCP_HTTP_PORT` is unset — that's the normal stdio case.

---

## Transport Dispatch (`main.go`)

`main()` dispatches after loading config and creating the client:

```go
if cfg.HTTP.Port != "" {
    runHTTP(cfg, client, regCh, log)
} else {
    runStdio(client, regCh, log)
}
```

The existing `processMCPRequests()` is renamed `runStdio()` — no logic changes.

---

## HTTP Transport (`main.go` — new `runHTTP()`)

### Endpoint

Single endpoint: `POST /mcp`

All JSON-RPC traffic goes through this one path, per the MCP Streamable HTTP spec (2025-03-26).

### Request handling

1. Reject non-POST methods on `/mcp`: `GET /mcp` → `405 Method Not Allowed` (spec-required); all other paths → `404 Not Found`
2. Decode JSON-RPC request body
3. Validate `Content-Type: application/json` and `JSONRPC: "2.0"`
4. Route to the **same handler functions** used by stdio: `handleInitialize`, `handleToolsList`, `handleToolCall`
5. Collect response and write `Content-Type: application/json` + JSON body

### Response format

Plain JSON only — no SSE streaming. Per spec, the server may choose between `application/json` and `text/event-stream`; we choose JSON because:
- All operations (`initialize`, `tools/list`, `tools/call`) are single request→response
- Open WebUI must support plain JSON per spec
- SSE adds ~200 lines of complexity with no current benefit

### Response writers

The current `writeResponse()` and `writeError()` write to stdout. HTTP gets dedicated variants:

```go
func httpWriteResponse(w http.ResponseWriter, resp MCPResponse)
func httpWriteError(w http.ResponseWriter, id interface{}, code, message string)
```

Same JSON structure, different sink. The shared handler functions (`handleInitialize` etc.) are refactored to return `MCPResponse` instead of calling `writeResponse()` directly, so both transports can use them.

### Notifications

`notifications/initialized` → `202 Accepted`, empty body (same as spec requires for notification-only POST bodies).

### Security

No origin validation — Docker network isolation is sufficient for local use. No authentication.

### Startup log

```
INFO Starting Tally MCP server in HTTP mode on 0.0.0.0:9090
```

---

## Protocol Version Fix

Both transports advertise `"protocolVersion": "2024-11-05"` in the `initialize` response.

The current value `"2025-11-25"` is not a published MCP spec version and may cause clients (including Open WebUI) to reject the handshake.

---

## Shared Handler Refactor

Currently `handleInitialize`, `handleToolsList`, `handleToolCall` each call `writeResponse()` directly. To share them across transports, they are refactored to **return** `MCPResponse`:

```go
func buildInitializeResponse(req MCPRequest) MCPResponse
func buildToolsListResponse(req MCPRequest, h *mcp.Handler) MCPResponse
func buildToolCallResponse(req MCPRequest, h *mcp.Handler, log *logger.Logger) (MCPResponse, error)
```

stdio calls `writeResponse(buildInitializeResponse(req))`.  
HTTP calls `httpWriteResponse(w, buildInitializeResponse(req))`.

No logic duplication.

---

## Docker Usage

No Dockerfile changes required. To run in HTTP mode:

```yaml
# docker-compose.yml
services:
  tally-mcp:
    image: tally-mcp
    environment:
      - MCP_HTTP_PORT=9090
      - MCP_HTTP_HOST=0.0.0.0
      - TALLY_HOST=4.186.35.209
      - TALLY_PORT=9000
      - TALLY_COMPANY=Dalade Private Limited
    ports:
      - "9090:9090"
```

Open WebUI: Settings → Admin → Tools → add `http://<host>:9090/mcp`

---

## Files Changed

| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `HTTPConfig` struct; load `MCP_HTTP_PORT`, `MCP_HTTP_HOST` |
| `pkg/config/config_test.go` | Add tests for HTTP config loading |
| `main.go` | Rename `processMCPRequests` → `runStdio`; add `runHTTP`; refactor handlers to return `MCPResponse`; fix protocol version |

No new packages. No new files beyond the spec doc.

---

## Out of Scope

- SSE streaming responses
- Authentication / authorization
- Origin header validation
- TLS/HTTPS
- Session management (`Mcp-Session-Id`)
- `GET /mcp` SSE stream (returns 405)
