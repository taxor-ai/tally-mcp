package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/mcp"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

// Type aliases for test convenience
type MCPRequest = mcp.JSONRPCRequest
type MCPResponse = mcp.JSONRPCResponse
type ToolResult = mcp.ToolResult
type ContentBlock = mcp.ContentBlock

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
	setupTemplatesDir(t)

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

	// Load registry
	registry, err := tally.LoadRegistry("tools")
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Create MCP handler
	handler := mcp.NewHandler(client, registry, log)

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
	companies, ok := resultMap["companies"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected companies to be []map[string]interface{}, got %T", resultMap["companies"])
	}

	// Verify companies are not empty
	if len(companies) == 0 {
		t.Fatal("Expected at least one company in response")
	}

	// Verify company structure
	for i, company := range companies {
		if company["name"] == "" || company["name"] == nil {
			t.Errorf("Company %d has empty name", i)
		}
		if company["guid"] == "" || company["guid"] == nil {
			t.Errorf("Company %d has empty guid", i)
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
	setupTemplatesDir(t)

	// Create logger
	log, err := logger.New("debug", "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create Tally client
	client := tally.NewClient("localhost", 9900, 30)
	client.SetCompany("TestCompany")

	// Load registry
	registry, err := tally.LoadRegistry("tools")
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Create MCP handler
	handler := mcp.NewHandler(client, registry, log)

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
	setupTemplatesDir(t)

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
	registry, err := tally.LoadRegistry("tools")
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}
	handler := mcp.NewHandler(client, registry, log)

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
	setupTemplatesDir(t)

	// Create logger
	log, _ := logger.New("warn", "")

	// Create client
	client := tally.NewClient("localhost", 9900, 30)
	client.SetCompany("TestCompany")

	// Load registry
	registry, err := tally.LoadRegistry("tools")
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Create handler
	handler := mcp.NewHandler(client, registry, log)

	// Get list of tools
	tools := handler.ListTools()

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
	// Set up templates directory (use os.Setenv since benchmarks don't use t.Cleanup)
	os.Setenv("TALLY_TOOLS_DIR", "tools")
	defer os.Unsetenv("TALLY_TOOLS_DIR")

	// Create logger (use /dev/null equivalent)
	log, _ := logger.New("warn", "")

	// Create handler
	client := tally.NewClient("localhost", 9900, 30)
	client.SetCompany("TestCompany")
	registry, _ := tally.LoadRegistry("tools")
	handler := mcp.NewHandler(client, registry, log)

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.HandleToolCall("get_companies", map[string]interface{}{})
	}
}

// setupTemplatesDir sets the TALLY_TOOLS_DIR environment variable for tests
// Finds the templates directory relative to the project root
func setupTemplatesDir(t *testing.T) {
	// Try multiple paths to find templates from different test working directories
	candidates := []string{
		"tools",
		"../../tools",
		"../../../tools",
	}

	var found string
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			found = candidate
			break
		}
	}

	if found == "" {
		t.Fatal("Could not find tools from any expected path")
	}

	os.Setenv("TALLY_TOOLS_DIR", found)
	t.Cleanup(func() { os.Unsetenv("TALLY_TOOLS_DIR") })
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

// TestNewToolsDiscovery verifies that the new voucher tools are discoverable
func TestNewToolsDiscovery(t *testing.T) {
	setupTemplatesDir(t)

	// Load registry
	registry, err := tally.LoadRegistry("tools")
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Expected new tools
	expectedTools := []string{
		"get_sales_vouchers",
		"create_journal_voucher",
		"create_sales_voucher",
		"get_payments",
		"create_payment",
	}

	// Get all tools from registry
	allTools := registry.All()
	toolNames := make(map[string]bool)
	for _, tool := range allTools {
		toolNames[tool.Name] = true
	}

	// Verify each expected tool exists
	for _, toolName := range expectedTools {
		if !toolNames[toolName] {
			t.Errorf("Expected tool %q not found in registry", toolName)
		}
	}

	// Verify we have at least the new tools (plus existing ones)
	if len(allTools) < len(expectedTools) {
		t.Errorf("Expected at least %d tools, got %d", len(expectedTools), len(allTools))
	}

	t.Logf("✓ All new tools discovered: %v", expectedTools)
}

// TestNewToolsSchemas verifies that new tools have correct input schemas
func TestNewToolsSchemas(t *testing.T) {
	setupTemplatesDir(t)

	registry, err := tally.LoadRegistry("tools")
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	testCases := []struct {
		toolName       string
		requiredFields []string
	}{
		{
			toolName:       "get_sales_vouchers",
			requiredFields: []string{"party_ledger_name"},
		},
		{
			toolName:       "create_journal_voucher",
			requiredFields: []string{"date", "reference", "narration", "lines"},
		},
		{
			toolName:       "create_sales_voucher",
			requiredFields: []string{"date", "reference", "narration", "party_ledger_name", "lines"},
		},
		{
			toolName:       "get_payments",
			requiredFields: []string{"party_ledger_name"},
		},
		{
			toolName:       "create_payment",
			requiredFields: []string{"date", "reference", "narration", "party_ledger_name", "lines"},
		},
	}

	for _, tc := range testCases {
		tool := registry.Get(tc.toolName)
		if tool == nil {
			t.Errorf("Tool %q not found", tc.toolName)
			continue
		}

		// Verify required fields exist in schema
		for _, field := range tc.requiredFields {
			if !containsString(tool.InputSchema.Required, field) {
				t.Errorf("Tool %q missing required field %q", tc.toolName, field)
			}
		}

		t.Logf("✓ Tool %q has correct schema", tc.toolName)
	}
}

// Helper function
func containsString(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
