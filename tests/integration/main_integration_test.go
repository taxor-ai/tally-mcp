// +build integration

package main

import (
	"os"
	"strconv"
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
func getIntEnvOrDefault(key string, defaultVal int) int {
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

// TestGetCompaniesRealTally tests get_companies against a real Tally instance
// Run with: go test -tags=integration ./...
// Requires: TALLY_HOST, TALLY_COMPANY, TALLY_PORT environment variables
func TestGetCompaniesRealTally(t *testing.T) {
	// Get Tally connection details from environment
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Fatal("TALLY_HOST environment variable is required")
	}

	company := os.Getenv("TALLY_COMPANY")
	if company == "" {
		t.Fatal("TALLY_COMPANY environment variable is required")
	}

	port := getIntEnvOrDefault("TALLY_PORT", 9900)

	// Create logger
	log, err := logger.New("debug", "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create Tally client
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)

	// Test Ping connectivity first
	err = client.Ping()
	if err != nil {
		t.Fatalf("Failed to connect to Tally at %s: %v", host, err)
	}
	t.Logf("✓ Connected to Tally at %s", host)

	// Create MCP handler
	handler := mcp.NewHandler(client, log)

	// Call get_companies
	result, err := handler.HandleToolCall("get_companies", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_companies failed: %v", err)
	}

	// Validate response structure
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	// Verify success field
	success, ok := resultMap["success"].(bool)
	if !ok || !success {
		t.Fatal("Expected success to be true")
	}

	// Verify companies field
	companies, ok := resultMap["companies"].([]tally.Company)
	if !ok {
		t.Fatalf("Expected []tally.Company, got %T", resultMap["companies"])
	}

	if len(companies) == 0 {
		t.Fatal("Expected at least one company from Tally")
	}

	// Verify company data
	for i, company := range companies {
		if company.Name == "" {
			t.Errorf("Company %d has empty Name", i)
		}
		if company.GUID == "" {
			t.Errorf("Company %d has empty GUID", i)
		}
		t.Logf("  Company %d: %s (%s)", i+1, company.Name, company.GUID)
	}

	t.Logf("✓ get_companies returned %d companies from Tally", len(companies))
}

// TestExecuteTemplate tests the ExecuteTemplate method directly
func TestExecuteTemplate(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Fatal("TALLY_HOST environment variable is required")
	}

	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	client := tally.NewClient(host, port, 30)
	client.SetCompany(os.Getenv("TALLY_COMPANY"))

	// Test Ping first
	err := client.Ping()
	if err != nil {
		t.Fatalf("Failed to connect to Tally: %v", err)
	}

	// Execute the template
	xmlResponse, err := client.ExecuteTemplate("company/get_companies", map[string]string{})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	if len(xmlResponse) == 0 {
		t.Fatal("Expected non-empty XML response from Tally")
	}

	t.Logf("✓ Received %d bytes of XML from Tally", len(xmlResponse))
	t.Logf("Response (first 500 chars): %s...", string(xmlResponse)[:min(len(xmlResponse), 500)])
}

// TestParseCompaniesResponse tests parsing of Tally's XML response
func TestParseCompaniesResponse(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Fatal("TALLY_HOST environment variable is required")
	}

	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	client := tally.NewClient(host, port, 30)
	client.SetCompany(os.Getenv("TALLY_COMPANY"))

	// Test Ping first
	err := client.Ping()
	if err != nil {
		t.Fatalf("Failed to connect to Tally: %v", err)
	}

	// Get raw XML response
	xmlResponse, err := client.ExecuteTemplate("company/get_companies", map[string]string{})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	// Parse the response
	companies, err := tally.ParseCompaniesResponse(xmlResponse)
	if err != nil {
		t.Fatalf("ParseCompaniesResponse failed: %v", err)
	}

	if len(companies) == 0 {
		t.Fatal("Expected at least one company")
	}

	// Verify parsed data
	for i, company := range companies {
		if company.Name == "" {
			t.Errorf("Company %d has empty Name", i)
		}
		if company.GUID == "" {
			t.Errorf("Company %d has empty GUID", i)
		}
		t.Logf("  Parsed Company %d: %s (%s)", i+1, company.Name, company.GUID)
	}

	t.Logf("✓ Successfully parsed %d companies from Tally response", len(companies))
}

// TestTallyConnectionError tests error handling for connection failures
func TestTallyConnectionError(t *testing.T) {
	// Create client with invalid host
	client := tally.NewClient("invalid-host-12345", 9900, 1) // 1 second timeout

	// Try to ping - should fail
	err := client.Ping()
	if err == nil {
		t.Fatal("Expected connection error for invalid host")
	}

	t.Logf("✓ Connection error handled correctly: %v", err)
}

// TestCompleteIntegrationFlow tests the complete flow from request to response
func TestCompleteIntegrationFlow(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Fatal("TALLY_HOST environment variable is required")
	}

	company := os.Getenv("TALLY_COMPANY")
	if company == "" {
		t.Fatal("TALLY_COMPANY environment variable is required")
	}

	// Create logger
	log, err := logger.New("info", "")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create client
	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)

	// Verify connectivity
	err = client.Ping()
	if err != nil {
		t.Fatalf("Failed to connect to Tally: %v", err)
	}

	// Create handler
	handler := mcp.NewHandler(client, log)

	// Simulate MCP request for get_companies
	result, err := handler.HandleToolCall("get_companies", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Tool call failed: %v", err)
	}

	// Verify result
	resultMap := result.(map[string]interface{})
	companies := resultMap["companies"].([]tally.Company)

	if len(companies) == 0 {
		t.Fatal("No companies returned from Tally")
	}

	t.Logf("✓ Complete integration flow successful")
	t.Logf("  Connected to: %s (company: %s)", host, company)
	t.Logf("  Retrieved: %d companies", len(companies))
	for i, c := range companies {
		t.Logf("    %d. %s", i+1, c.Name)
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestGetLedgersIntegration tests the get_ledgers tool
func TestGetLedgersIntegration(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Skip("TALLY_HOST not set")
	}

	log, _ := logger.New("warn", "")
	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)
	handler := mcp.NewHandler(client, log)

	result, err := handler.HandleToolCall("get_ledgers", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_ledgers failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Fatal("get_ledgers returned success=false")
	}

	ledgers := resultMap["ledgers"].([]tally.Ledger)
	t.Logf("✓ get_ledgers returned %d ledgers", len(ledgers))
	for i, ledger := range ledgers {
		if i < 3 {
			t.Logf("  Ledger %d: %s (Parent: %s)", i+1, ledger.Name, ledger.ParentGroup)
		}
	}
}

// TestGetDebtorsIntegration tests the get_debtors tool
func TestGetDebtorsIntegration(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Skip("TALLY_HOST not set")
	}

	log, _ := logger.New("warn", "")
	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)
	handler := mcp.NewHandler(client, log)

	result, err := handler.HandleToolCall("get_debtors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_debtors failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Fatal("get_debtors returned success=false")
	}

	debtors := resultMap["debtors"].([]tally.Debtor)
	t.Logf("✓ get_debtors returned %d debtors", len(debtors))
	for i, debtor := range debtors {
		if i < 3 {
			t.Logf("  Debtor %d: %s (Outstanding: %.2f)", i+1, debtor.Name, debtor.OutstandingAmount)
		}
	}
}

// TestGetCreditorsIntegration tests the get_creditors tool
func TestGetCreditorsIntegration(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Skip("TALLY_HOST not set")
	}

	log, _ := logger.New("warn", "")
	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)
	handler := mcp.NewHandler(client, log)

	result, err := handler.HandleToolCall("get_creditors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_creditors failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Fatal("get_creditors returned success=false")
	}

	creditors := resultMap["creditors"].([]tally.Creditor)
	t.Logf("✓ get_creditors returned %d creditors", len(creditors))
	for i, creditor := range creditors {
		if i < 3 {
			t.Logf("  Creditor %d: %s (Outstanding: %.2f)", i+1, creditor.Name, creditor.OutstandingAmount)
		}
	}
}

// TestGetVouchersIntegration tests the get_vouchers tool
func TestGetVouchersIntegration(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Skip("TALLY_HOST not set")
	}

	log, _ := logger.New("warn", "")
	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)
	handler := mcp.NewHandler(client, log)

	result, err := handler.HandleToolCall("get_vouchers", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_vouchers failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Fatal("get_vouchers returned success=false")
	}

	vouchers := resultMap["vouchers"].([]tally.Voucher)
	t.Logf("✓ get_vouchers returned %d vouchers", len(vouchers))
	for i, voucher := range vouchers {
		if i < 3 {
			t.Logf("  Voucher %d: %s (%s) - %s Amount: %.2f", i+1, voucher.VoucherID, voucher.Type, voucher.Date, voucher.Amount)
		}
	}
}

// TestAllGetToolsSequenceIntegration tests all get_ tools in sequence
func TestAllGetToolsSequenceIntegration(t *testing.T) {
	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Skip("TALLY_HOST not set")
	}

	log, _ := logger.New("warn", "")
	port := getIntEnvOrDefault("TALLY_PORT", 9900)
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)
	handler := mcp.NewHandler(client, log)

	t.Logf("Testing all GET tools sequentially...")

	// Test 1: get_companies
	result, err := handler.HandleToolCall("get_companies", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_companies failed: %v", err)
	}
	companies := result.(map[string]interface{})["companies"].([]tally.Company)
	t.Logf("✓ get_companies: %d companies", len(companies))

	// Test 2: get_ledgers
	result, err = handler.HandleToolCall("get_ledgers", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_ledgers failed: %v", err)
	}
	ledgers := result.(map[string]interface{})["ledgers"].([]tally.Ledger)
	t.Logf("✓ get_ledgers: %d ledgers", len(ledgers))

	// Test 3: get_debtors
	result, err = handler.HandleToolCall("get_debtors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_debtors failed: %v", err)
	}
	debtors := result.(map[string]interface{})["debtors"].([]tally.Debtor)
	t.Logf("✓ get_debtors: %d debtors", len(debtors))

	// Test 4: get_creditors
	result, err = handler.HandleToolCall("get_creditors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_creditors failed: %v", err)
	}
	creditors := result.(map[string]interface{})["creditors"].([]tally.Creditor)
	t.Logf("✓ get_creditors: %d creditors", len(creditors))

	// Test 5: get_vouchers
	result, err = handler.HandleToolCall("get_vouchers", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_vouchers failed: %v", err)
	}
	vouchers := result.(map[string]interface{})["vouchers"].([]tally.Voucher)
	t.Logf("✓ get_vouchers: %d vouchers", len(vouchers))

	t.Logf("\n✓✓✓ All 5 GET tools working successfully ✓✓✓")
	t.Logf("Summary: Companies=%d, Ledgers=%d, Debtors=%d, Creditors=%d, Vouchers=%d",
		len(companies), len(ledgers), len(debtors), len(creditors), len(vouchers))
}
