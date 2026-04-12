package mcp

// JSONRPCRequest represents a JSON-RPC 2.0 request for tool calling
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
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
	Type string `json:"type"`
	Text string `json:"text"`
}
