package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/taxor-ai/tally-mcp/config"
	"github.com/taxor-ai/tally-mcp/logger"
	"github.com/taxor-ai/tally-mcp/mcp"
	"github.com/taxor-ai/tally-mcp/tally"
)

// MCPRequest represents a JSON-RPC 2.0 request for tool calling
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ToolResult is the response format for tool calls
type ToolResult struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a single content block in the result
type ContentBlock struct {
	Type string      `json:"type"`
	Text string      `json:"text"`
}

func main() {
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
			// Flush any pending logs
			_ = log.Sync()
		}
	}()

	log.Infof("Starting Tally MCP server (version 0.1.0)")
	log.Infof("Tally host: %s:%d, company: %s", cfg.Tally.Host, cfg.Tally.Port, cfg.Tally.Company)

	// Create Tally client
	client := tally.NewClient(cfg.Tally.Host, cfg.Tally.Port, 30)
	client.SetCompany(cfg.Tally.Company)

	// Create MCP handler
	handler := mcp.NewHandler(client, log)

	// Process requests from stdin
	reader := bufio.NewReader(os.Stdin)
	processMCPRequests(reader, handler, log)
}

// processMCPRequests reads JSON-RPC requests from stdin and processes them
func processMCPRequests(reader *bufio.Reader, handler *mcp.Handler, log *logger.Logger) {
	decoder := json.NewDecoder(reader)

	for {
		var req MCPRequest
		err := decoder.Decode(&req)

		// Check for EOF
		if err == io.EOF {
			log.Infof("Client disconnected")
			break
		}

		if err != nil {
			log.Warnf("Error decoding request: %v", err)
			writeError(nil, "parse_error", fmt.Sprintf("Invalid request: %v", err))
			continue
		}

		// Validate request
		if req.JSONRPC != "2.0" {
			log.Warnf("Invalid JSONRPC version: %s", req.JSONRPC)
			writeError(req.ID, "invalid_request", "JSONRPC version must be 2.0")
			continue
		}

		// Handle different methods
		switch req.Method {
		case "tools/call":
			handleToolCall(req, handler, log)

		case "tools/list":
			handleToolsList(req, handler, log)

		default:
			log.Warnf("Unknown method: %s", req.Method)
			writeError(req.ID, "method_not_found", fmt.Sprintf("Method %s not found", req.Method))
		}
	}
}

// handleToolCall processes tool call requests
func handleToolCall(req MCPRequest, handler *mcp.Handler, log *logger.Logger) {
	toolName, ok := req.Params["name"].(string)
	if !ok {
		log.Warnf("Missing or invalid tool name in request")
		writeError(req.ID, "invalid_params", "Tool name must be a string")
		return
	}

	arguments, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	log.Debugf("Calling tool: %s with arguments: %v", toolName, arguments)

	// Call handler
	result, err := handler.HandleToolCall(toolName, arguments)
	if err != nil {
		log.Warnf("Tool call failed: %v", err)
		writeError(req.ID, "tool_error", fmt.Sprintf("Tool execution failed: %v", err))
		return
	}

	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Warnf("Error marshaling result: %v", err)
		writeError(req.ID, "result_error", "Failed to marshal result")
		return
	}

	// Format response
	contentBlock := ContentBlock{
		Type: "text",
		Text: string(resultJSON),
	}

	toolResult := ToolResult{
		Content: []ContentBlock{contentBlock},
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolResult,
	}

	writeResponse(response)
}

// handleToolsList returns the list of available tools
func handleToolsList(req MCPRequest, handler *mcp.Handler, log *logger.Logger) {
	log.Debugf("Listing available tools")

	tools := mcp.AllTools()

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}

	writeResponse(response)
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

// writeError writes a JSON-RPC error response to stdout
func writeError(id interface{}, code string, message string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	writeResponse(resp)
}
