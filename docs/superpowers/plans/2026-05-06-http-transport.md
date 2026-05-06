# HTTP Transport Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Streamable HTTP transport to tally-mcp so it works with Open WebUI, while keeping existing stdio transport fully intact.

**Architecture:** Transport is selected by presence of `MCP_HTTP_PORT` env var at startup. Both transports share the same `buildXxx` response-builder functions. The HTTP server waits for the tool registry to load before binding the port, then serves plain JSON responses on `POST /mcp`.

**Tech Stack:** Go standard library only — `net/http`, `encoding/json`. No new dependencies.

---

## File Map

| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `HTTPConfig` struct; load `MCP_HTTP_PORT`, `MCP_HTTP_HOST` in `Load()` |
| `pkg/config/config_test.go` | Add two tests: HTTP defaults and HTTP env vars |
| `main.go` | Fix protocol version; rename `processMCPRequests` → `runStdio`; refactor three `handleXxx` functions to `buildXxx` (return `MCPResponse` instead of calling `writeResponse`); add `buildErrorResponse`; add `runHTTP`, `serveMCP`, `httpWriteResponse`, `httpWriteError`; update `main()` dispatch |
| `main_test.go` | New file: unit test `buildInitializeResponse`; integration tests for `serveMCP` using `httptest` |

---

## Task 1: Add HTTPConfig to config

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

Add these two functions to the bottom of `pkg/config/config_test.go`:

```go
func TestHTTPConfigDefaults(t *testing.T) {
	os.Setenv("TALLY_HOST", "localhost")
	os.Setenv("TALLY_PORT", "9900")
	os.Setenv("TALLY_COMPANY", "DemoCompany")
	os.Unsetenv("MCP_HTTP_PORT")
	os.Unsetenv("MCP_HTTP_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.HTTP.Port != "" {
		t.Errorf("Expected empty HTTP port (stdio mode), got %q", cfg.HTTP.Port)
	}
	if cfg.HTTP.Host != "0.0.0.0" {
		t.Errorf("Expected default HTTP host '0.0.0.0', got %q", cfg.HTTP.Host)
	}
}

func TestHTTPConfigFromEnv(t *testing.T) {
	os.Setenv("TALLY_HOST", "localhost")
	os.Setenv("TALLY_PORT", "9900")
	os.Setenv("TALLY_COMPANY", "DemoCompany")
	os.Setenv("MCP_HTTP_PORT", "9090")
	os.Setenv("MCP_HTTP_HOST", "127.0.0.1")
	defer os.Unsetenv("MCP_HTTP_PORT")
	defer os.Unsetenv("MCP_HTTP_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.HTTP.Port != "9090" {
		t.Errorf("Expected HTTP port '9090', got %q", cfg.HTTP.Port)
	}
	if cfg.HTTP.Host != "127.0.0.1" {
		t.Errorf("Expected HTTP host '127.0.0.1', got %q", cfg.HTTP.Host)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go test ./pkg/config/... -v -run "TestHTTPConfig"
```

Expected: compile error — `cfg.HTTP` undefined.

- [ ] **Step 3: Add HTTPConfig to config.go**

In `pkg/config/config.go`, add the `HTTPConfig` struct and update `Config`:

```go
type HTTPConfig struct {
	Port string // empty = stdio mode; e.g. "9090"
	Host string // bind address, default "0.0.0.0"
}

type Config struct {
	Tally  TallyConfig
	Logger LoggerConfig
	HTTP   HTTPConfig
}
```

In `Load()`, add HTTP config population inside the `cfg := &Config{...}` block:

```go
cfg := &Config{
	Tally: TallyConfig{
		Host:    os.Getenv("TALLY_HOST"),
		Company: os.Getenv("TALLY_COMPANY"),
		Timeout: 30,
	},
	Logger: LoggerConfig{
		Level: getEnvOrDefault("TALLY_LOG_LEVEL", "info"),
		File:  os.Getenv("TALLY_LOG_FILE"),
	},
	HTTP: HTTPConfig{
		Port: os.Getenv("MCP_HTTP_PORT"),
		Host: getEnvOrDefault("MCP_HTTP_HOST", "0.0.0.0"),
	},
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./pkg/config/... -v
```

Expected: all config tests pass including the two new ones.

- [ ] **Step 5: Commit**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
git add pkg/config/config.go pkg/config/config_test.go
git commit -m "feat: add HTTPConfig to config for MCP_HTTP_PORT/MCP_HTTP_HOST"
```

---

## Task 2: Refactor handler functions + fix protocol version

**Files:**
- Create: `main_test.go`
- Modify: `main.go`

The three `handleXxx` functions currently call `writeResponse()` directly (stdout). Refactor them to return `MCPResponse` so both transports can share them. Also fix the invalid protocol version `"2025-11-25"` → `"2024-11-05"`.

- [ ] **Step 1: Create main_test.go with a failing test**

Create `/Users/sree/Projects/Branding/tally-mcp/main_test.go`:

```go
package main

import (
	"testing"
)

func TestBuildInitializeResponse(t *testing.T) {
	req := MCPRequest{JSONRPC: "2.0", ID: float64(1), Method: "initialize"}
	resp := buildInitializeResponse(req)

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got %q", resp.JSONRPC)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected Result to be map[string]interface{}, got %T", resp.Result)
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected serverInfo to be map[string]interface{}")
	}
	if serverInfo["name"] != "Tally MCP" {
		t.Errorf("Expected serverInfo.name 'Tally MCP', got %v", serverInfo["name"])
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go test . -v -run "TestBuildInitializeResponse"
```

Expected: compile error — `buildInitializeResponse` undefined.

- [ ] **Step 3: Refactor main.go — handler functions**

Replace the three `handleXxx` functions and `writeError` in `main.go` with these:

```go
// buildInitializeResponse builds the MCP initialize response.
func buildInitializeResponse(req MCPRequest) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "Tally MCP",
				"version": "0.1.0",
			},
		},
	}
}

// buildToolsListResponse builds the tools/list response.
func buildToolsListResponse(req MCPRequest, handler *mcp.Handler) MCPResponse {
	tools := handler.ListTools()
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// buildToolCallResponse dispatches a tool call and returns the response.
func buildToolCallResponse(req MCPRequest, handler *mcp.Handler, log *logger.Logger) MCPResponse {
	toolName, ok := req.Params["name"].(string)
	if !ok {
		log.Warnf("Missing or invalid tool name in request")
		return buildErrorResponse(req.ID, "invalid_params", "Tool name must be a string")
	}

	arguments, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	log.Debugf("Calling tool: %s with arguments: %v", toolName, arguments)

	result, err := handler.HandleToolCall(toolName, arguments)
	if err != nil {
		log.Warnf("Tool call failed: %v", err)
		return buildErrorResponse(req.ID, "tool_error", fmt.Sprintf("Tool execution failed: %v", err))
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Warnf("Error marshaling result: %v", err)
		return buildErrorResponse(req.ID, "result_error", "Failed to marshal result")
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolResult{
			Content: []ContentBlock{{
				Type: "text",
				Text: string(resultJSON),
			}},
		},
	}
}

// buildErrorResponse builds a JSON-RPC error response.
func buildErrorResponse(id interface{}, code, message string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

// writeResponse writes a JSON-RPC response to stdout.
func writeResponse(resp MCPResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// writeError writes a JSON-RPC error response to stdout.
func writeError(id interface{}, code string, message string) {
	writeResponse(buildErrorResponse(id, code, message))
}
```

- [ ] **Step 4: Rename processMCPRequests → runStdio and update call sites**

Replace `processMCPRequests` with `runStdio` in `main.go`. The function now creates its own reader (remove the `reader` parameter) and calls `writeResponse(buildXxx(...))`:

```go
// runStdio reads JSON-RPC requests from stdin and processes them.
// The handler is resolved lazily on first tools/list or tools/call,
// so initialize can respond immediately without waiting for registry load.
func runStdio(client *tally.Client, regCh <-chan registryResult, log *logger.Logger) {
	decoder := json.NewDecoder(bufio.NewReader(os.Stdin))
	var handler *mcp.Handler

	getHandler := func(reqID interface{}) *mcp.Handler {
		if handler != nil {
			return handler
		}
		result := <-regCh
		if result.err != nil {
			log.Warnf("Tool registry failed to load: %v", result.err)
			writeError(reqID, "internal_error", fmt.Sprintf("Tool registry unavailable: %v", result.err))
			return nil
		}
		log.Infof("Tool registry ready with %d tools", len(result.registry.All()))
		handler = mcp.NewHandler(client, result.registry, log)
		return handler
	}

	for {
		var req MCPRequest
		err := decoder.Decode(&req)
		if err == io.EOF {
			log.Infof("Client disconnected")
			break
		}
		if err != nil {
			log.Warnf("Error decoding request: %v", err)
			writeError(nil, "parse_error", fmt.Sprintf("Invalid request: %v", err))
			continue
		}

		if req.JSONRPC != "2.0" {
			log.Warnf("Invalid JSONRPC version: %s", req.JSONRPC)
			writeError(req.ID, "invalid_request", "JSONRPC version must be 2.0")
			continue
		}

		isNotification := req.ID == nil

		switch req.Method {
		case "initialize":
			log.Debugf("Handling initialize request")
			writeResponse(buildInitializeResponse(req))
		case "notifications/initialized":
			log.Debugf("Client initialized")
		case "tools/list":
			if h := getHandler(req.ID); h != nil {
				log.Debugf("Listing available tools")
				writeResponse(buildToolsListResponse(req, h))
			}
		case "tools/call":
			if h := getHandler(req.ID); h != nil {
				writeResponse(buildToolCallResponse(req, h, log))
			}
		default:
			if !isNotification {
				log.Warnf("Unknown method: %s", req.Method)
				writeError(req.ID, "method_not_found", fmt.Sprintf("Method %s not found", req.Method))
			}
		}
	}
}
```

Update `main()` to call `runStdio` without the reader argument and without the intermediate `reader` variable:

```go
// Replace this:
reader := bufio.NewReader(os.Stdin)
processMCPRequests(reader, client, regCh, log)

// With this:
runStdio(client, regCh, log)
```

- [ ] **Step 5: Run tests to confirm they pass**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go test . -v -run "TestBuildInitializeResponse"
```

Expected: PASS.

- [ ] **Step 6: Run all tests**

```bash
go test ./... -v
```

Expected: all existing tests pass.

- [ ] **Step 7: Commit**

```bash
git add main.go main_test.go
git commit -m "refactor: extract buildXxx response builders, fix protocol version to 2024-11-05"
```

---

## Task 3: Add HTTP transport

**Files:**
- Modify: `main_test.go` (add HTTP tests)
- Modify: `main.go` (add runHTTP, serveMCP, httpWriteResponse, httpWriteError; update main())

- [ ] **Step 1: Replace main_test.go with the complete file**

Task 2 created `main_test.go` with only `"testing"` imported. Replace it now with the full file that includes all imports and all tests (the HTTP tests need `net/http`, `net/http/httptest`, `strings`):

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
)

func TestBuildInitializeResponse(t *testing.T) {
	req := MCPRequest{JSONRPC: "2.0", ID: float64(1), Method: "initialize"}
	resp := buildInitializeResponse(req)

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got %q", resp.JSONRPC)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected Result to be map[string]interface{}, got %T", resp.Result)
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected serverInfo to be map[string]interface{}")
	}
	if serverInfo["name"] != "Tally MCP" {
		t.Errorf("Expected serverInfo.name 'Tally MCP', got %v", serverInfo["name"])
	}
}

func TestServeMCPGETReturns405(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/mcp")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", resp.StatusCode)
	}
}

func TestServeMCPInitialize(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %q", ct)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	inner, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'result' field in response, got: %v", result)
	}
	if inner["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", inner["protocolVersion"])
	}
}

func TestServeMCPNotificationsInitialized(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected 202, got %d", resp.StatusCode)
	}
}

func TestServeMCPUnknownMethod(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"unknown/method","params":{}}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if _, hasError := result["error"]; !hasError {
		t.Errorf("Expected 'error' field in response for unknown method, got: %v", result)
	}
}

func TestServeMCPInvalidJSON(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{invalid json`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if _, hasError := result["error"]; !hasError {
		t.Errorf("Expected 'error' field for invalid JSON, got: %v", result)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go test . -v -run "TestServeMCP"
```

Expected: compile error — `serveMCP` undefined.

- [ ] **Step 3: Add HTTP transport functions to main.go**

Add `"net/http"` to the import block at the top of `main.go`. The full import block should be:

```go
import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/taxor-ai/tally-mcp/pkg/config"
	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/mcp"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)
```

Add these three functions to `main.go` (after `writeError`):

```go
// runHTTP starts the MCP HTTP server. Blocks until the registry is ready, then binds the port.
func runHTTP(cfg *config.Config, client *tally.Client, regCh <-chan registryResult, log *logger.Logger) {
	result := <-regCh
	if result.err != nil {
		log.Errorf("Tool registry failed to load: %v", result.err)
		os.Exit(1)
	}
	log.Infof("Tool registry ready with %d tools", len(result.registry.All()))
	handler := mcp.NewHandler(client, result.registry, log)

	addr := fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	log.Infof("Starting Tally MCP server in HTTP mode on %s", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, handler, log)
	})

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		os.Exit(1)
	}
}

// serveMCP handles a single HTTP request to the /mcp endpoint.
func serveMCP(w http.ResponseWriter, r *http.Request, handler *mcp.Handler, log *logger.Logger) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpWriteError(w, nil, "parse_error", fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.JSONRPC != "2.0" {
		httpWriteError(w, req.ID, "invalid_request", "JSONRPC version must be 2.0")
		return
	}

	switch req.Method {
	case "initialize":
		log.Debugf("HTTP: handling initialize request")
		httpWriteResponse(w, buildInitializeResponse(req))
	case "notifications/initialized":
		log.Debugf("HTTP: client initialized")
		w.WriteHeader(http.StatusAccepted)
	case "tools/list":
		log.Debugf("HTTP: listing tools")
		httpWriteResponse(w, buildToolsListResponse(req, handler))
	case "tools/call":
		httpWriteResponse(w, buildToolCallResponse(req, handler, log))
	default:
		log.Warnf("HTTP: unknown method: %s", req.Method)
		httpWriteError(w, req.ID, "method_not_found", fmt.Sprintf("Method %s not found", req.Method))
	}
}

// httpWriteResponse writes a JSON-RPC response as application/json to an HTTP response.
func httpWriteResponse(w http.ResponseWriter, resp MCPResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// httpWriteError writes a JSON-RPC error response to an HTTP response.
func httpWriteError(w http.ResponseWriter, id interface{}, code, message string) {
	httpWriteResponse(w, buildErrorResponse(id, code, message))
}
```

- [ ] **Step 4: Update main() to dispatch on HTTP port**

Replace the end of `main()` — currently:

```go
reader := bufio.NewReader(os.Stdin)
processMCPRequests(reader, client, regCh, log)
```

With:

```go
if cfg.HTTP.Port != "" {
	runHTTP(cfg, client, regCh, log)
} else {
	runStdio(client, regCh, log)
}
```

- [ ] **Step 5: Run HTTP tests to confirm they pass**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go test . -v -run "TestServeMCP"
```

Expected: all five `TestServeMCP*` tests PASS.

- [ ] **Step 6: Run all tests**

```bash
go test ./... -v
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
git add main.go main_test.go
git commit -m "feat: add HTTP transport — POST /mcp serves JSON-RPC, transport selected by MCP_HTTP_PORT"
```

---

## Task 4: Smoke test

Verify the server works end-to-end with curl before connecting Open WebUI.

- [ ] **Step 1: Build**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go build -o tally-mcp .
```

Expected: exits 0, no errors.

- [ ] **Step 2: Start the server in HTTP mode**

In a separate terminal (or background), run:

```bash
MCP_HTTP_PORT=9090 \
TALLY_HOST=4.186.35.209 \
TALLY_PORT=9000 \
TALLY_COMPANY="Dalade Private Limited" \
TALLY_LOG_LEVEL=debug \
./tally-mcp
```

Expected log line:
```
Starting Tally MCP server in HTTP mode on 0.0.0.0:9090
```

- [ ] **Step 3: Test initialize**

```bash
curl -s -X POST http://localhost:9090/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | jq .
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": { "tools": {} },
    "serverInfo": { "name": "Tally MCP", "version": "0.1.0" }
  }
}
```

- [ ] **Step 4: Confirm GET returns 405**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X GET http://localhost:9090/mcp
```

Expected: `405`

- [ ] **Step 5: Test tools/list**

```bash
curl -s -X POST http://localhost:9090/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | jq '.result.tools | length'
```

Expected: a number > 0 (the count of loaded tools).

- [ ] **Step 6: Verify stdio mode is unchanged**

Kill the HTTP server. Run without `MCP_HTTP_PORT`:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | \
  TALLY_HOST=4.186.35.209 TALLY_PORT=9000 TALLY_COMPANY="Dalade Private Limited" \
  ./tally-mcp
```

Expected: JSON response on stdout, same as before. No HTTP binding.
