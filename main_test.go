package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/taxor-ai/tally-mcp/logger"
	"github.com/taxor-ai/tally-mcp/mcp"
	"github.com/taxor-ai/tally-mcp/tally"
)

// Helper to get int from environment with default
func getIntEnvOrDefaultMain(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// TestGetCompaniesUnit tests get_companies tool through the MCP handler
// This test skips if Tally is not available (requires TALLY_HOST env var)
func TestGetCompaniesUnit(t *testing.T) {
	// Skip if Tally is not available in unit test environment
	if os.Getenv("TALLY_HOST") == "" {
		t.Skip("TALLY_HOST not set, skipping (use integration test with -tags=integration)")
	}

	// Create logger
	log, err := logger.New("debug", "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create Tally client
	host := os.Getenv("TALLY_HOST")
	company := os.Getenv("TALLY_COMPANY")
	port := getIntEnvOrDefaultMain("TALLY_PORT", 9900)
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)

	// Create MCP handler
	handler := mcp.NewHandler(client, log)

	// Simulate processing the request
	toolName := "get_companies"
	arguments := map[string]interface{}{}

	result, err := handler.HandleToolCall(toolName, arguments)
	if err != nil {
		t.Fatalf("Failed to call tool: %v", err)
	}

	// Validate response structure
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Assert result is a map
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be map[string]interface{}, got %T", result)
	}

	// Verify success field
	success, ok := resultMap["success"].(bool)
	if !ok || !success {
		t.Fatal("Expected success to be true")
	}

	// Verify companies field exists
	companies, ok := resultMap["companies"].([]tally.Company)
	if !ok {
		t.Fatalf("Expected companies to be []tally.Company, got %T", resultMap["companies"])
	}

	// Verify companies are not empty
	if len(companies) == 0 {
		t.Fatal("Expected at least one company in response")
	}

	// Verify company structure
	for i, company := range companies {
		if company.Name == "" {
			t.Errorf("Company %d has empty Name", i)
		}
		if company.GUID == "" {
			t.Errorf("Company %d has empty GUID", i)
		}
	}

	t.Logf("✓ get_companies tool call succeeded")
	t.Logf("✓ Response contains %d companies", len(companies))
}

// TestMCPProtocolRequestProcessing tests the JSON-RPC request/response flow without calling Tally
func TestMCPProtocolRequestProcessing(t *testing.T) {
	// Create a test request
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "get_companies",
			"arguments": map[string]interface{}{},
		},
	}

	// Validate request structure
	if req.JSONRPC != "2.0" {
		t.Fatal("JSONRPC should be 2.0")
	}

	if req.Method != "tools/call" {
		t.Fatal("Method should be tools/call")
	}

	// Verify params can be extracted
	toolName, ok := req.Params["name"].(string)
	if !ok || toolName != "get_companies" {
		t.Fatal("Could not extract tool name from params")
	}

	arguments, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	if arguments == nil {
		t.Fatal("Arguments should not be nil")
	}

	t.Logf("✓ MCP request structure is valid")
}

// TestMCPResponseFormatting tests that responses are properly formatted
func TestMCPResponseFormatting(t *testing.T) {
	// Create a tool result
	toolResult := ToolResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: `{"success":true,"companies":[{"Name":"TestCo","GUID":"test001"}]}`,
			},
		},
	}

	// Create MCP response
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  toolResult,
	}

	// Marshal to JSON
	respJSON, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify JSON is valid
	var unmarshed MCPResponse
	err = json.Unmarshal(respJSON, &unmarshed)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify fields are preserved
	if unmarshed.JSONRPC != "2.0" {
		t.Fatal("JSONRPC not preserved")
	}

	if unmarshed.ID != float64(1) {
		t.Fatal("ID not preserved (JSON unmarshals to float64)")
	}

	t.Logf("✓ MCP response formatting is valid")
	t.Logf("Response JSON: %s", string(respJSON))
}

// TestUnknownToolErrorHandling tests that unknown tools return proper errors
func TestUnknownToolErrorHandling(t *testing.T) {
	// Create logger
	log, err := logger.New("debug", "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create Tally client
	client := tally.NewClient("localhost", 9900, 30)
	client.SetCompany("TestCompany")

	// Create MCP handler
	handler := mcp.NewHandler(client, log)

	// Try to call an unknown tool
	_, err = handler.HandleToolCall("unknown_tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for unknown tool")
	}

	if !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("Expected 'unknown tool' in error message, got: %v", err)
	}

	t.Logf("✓ Unknown tool error handling is correct")
}

// TestCompleteRequestResponse tests a complete request/response cycle
// Skips if Tally is not available (TALLY_HOST not set)
func TestCompleteRequestResponse(t *testing.T) {
	// Skip if Tally is not available
	if os.Getenv("TALLY_HOST") == "" {
		t.Skip("TALLY_HOST not set, skipping test that requires Tally")
	}

	// Create logger
	log, err := logger.New("info", "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create handler
	host := os.Getenv("TALLY_HOST")
	company := os.Getenv("TALLY_COMPANY")
	port := getIntEnvOrDefaultMain("TALLY_PORT", 9900)
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)
	handler := mcp.NewHandler(client, log)

	// Create a request JSON string (simulating what stdin would provide)
	requestJSON := `{
		"jsonrpc": "2.0",
		"id": 42,
		"method": "tools/call",
		"params": {
			"name": "get_companies",
			"arguments": {}
		}
	}`

	// Parse request
	var req MCPRequest
	err = json.Unmarshal([]byte(requestJSON), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	// Call the tool
	toolName := req.Params["name"].(string)
	arguments := req.Params["arguments"].(map[string]interface{})

	result, err := handler.HandleToolCall(toolName, arguments)
	if err != nil {
		t.Fatalf("Tool call failed: %v", err)
	}

	// Format response (same as handleToolCall in main.go)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolResult{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: string(resultJSON),
				},
			},
		},
	}

	// Marshal response
	respJSON, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify response structure
	var respUnmarshaled MCPResponse
	err = json.Unmarshal(respJSON, &respUnmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify ID matches
	if respUnmarshaled.ID != float64(42) {
		t.Fatalf("Expected ID 42, got %v", respUnmarshaled.ID)
	}

	t.Logf("✓ Complete request/response cycle successful")
	t.Logf("Request ID: %v, Response ID: %v", req.ID, respUnmarshaled.ID)
	t.Logf("Response: %s", string(respJSON))
}

// TestToolsList tests the tools/list method
func TestToolsList(t *testing.T) {
	// Get list of tools
	tools := mcp.AllTools()

	// Verify tools are available
	if len(tools) == 0 {
		t.Fatal("Expected at least one tool to be available")
	}

	// Verify get_companies tool exists
	found := false
	for _, tool := range tools {
		if tool.Name == "get_companies" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("get_companies tool not found in tools list")
	}

	t.Logf("✓ Tools list contains %d tools", len(tools))
	for _, tool := range tools {
		t.Logf("  - %s: %s", tool.Name, tool.Description)
	}
}

// BenchmarkGetCompanies benchmarks the get_companies tool performance
func BenchmarkGetCompanies(b *testing.B) {
	// Create logger (use /dev/null equivalent)
	log, _ := logger.New("warn", "")

	// Create handler
	client := tally.NewClient("localhost", 9900, 30)
	client.SetCompany("TestCompany")
	handler := mcp.NewHandler(client, log)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.HandleToolCall("get_companies", map[string]interface{}{})
	}
}

// Helper function to assert JSON is valid
func assertValidJSON(t *testing.T, data string) {
	var result interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
}

// Helper function to write JSON response to buffer (simulating stdout)
func writeResponseToBuffer(resp MCPResponse, buf *bytes.Buffer) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = buf.WriteString(string(data))
	return err
}
