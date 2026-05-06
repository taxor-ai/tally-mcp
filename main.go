package main

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

// Type aliases for convenience (actual types are in pkg/mcp)
type MCPRequest = mcp.JSONRPCRequest
type MCPResponse = mcp.JSONRPCResponse
type ToolResult = mcp.ToolResult
type ContentBlock = mcp.ContentBlock

type registryResult struct {
	registry *tally.Registry
	err      error
}

func main() {
	// Recover from panics and log to stderr
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "PANIC: %v\n", r)
			os.Exit(1)
		}
	}()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	log, err := logger.New(cfg.Logger.Level, cfg.Logger.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if log != nil {
			// Flush any pending logs (ignore errors)
			if err := log.Sync(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: error syncing logger: %v\n", err)
			}
		}
	}()

	log.Infof("Starting Tally MCP server (version 0.1.0)")
	log.Infof("Tally host: %s:%d, company: %s", cfg.Tally.Host, cfg.Tally.Port, cfg.Tally.Company)

	// Create Tally client
	client := tally.NewClient(cfg.Tally.Host, cfg.Tally.Port, 30)
	client.SetCompany(cfg.Tally.Company)

	// Determine tools directory
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining executable path: %v\n", err)
		os.Exit(1)
	}
	toolsDir := filepath.Join(filepath.Dir(exePath), "tools")

	// Load tool registry in background so we can respond to initialize immediately
	regCh := make(chan registryResult, 1)
	go func() {
		reg, err := tally.LoadRegistry(toolsDir)
		regCh <- registryResult{reg, err}
	}()

	// Dispatch to HTTP or stdio transport based on configuration
	if cfg.HTTP.Port != "" {
		runHTTP(cfg, client, regCh, log)
	} else {
		runStdio(client, regCh, log)
	}
}

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

// writeResponse writes a JSON-RPC response to stdout
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
		log.Errorf("HTTP server error: %v", err)
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
	if _, err := w.Write(data); err != nil {
		// Client disconnected; log but don't attempt another write
		fmt.Fprintf(os.Stderr, "Warning: failed to write HTTP response: %v\n", err)
	}
}

// httpWriteError writes a JSON-RPC error response to an HTTP response.
func httpWriteError(w http.ResponseWriter, id interface{}, code, message string) {
	httpWriteResponse(w, buildErrorResponse(id, code, message))
}
