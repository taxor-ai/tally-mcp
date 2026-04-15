//go:build integration

package main

import (
	"os"
	"strconv"
	"testing"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/mcp"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

// setupHandler creates a registry-backed MCP handler from environment variables.
// Tests skip if TALLY_HOST is not set.
func setupHandler(t *testing.T) *mcp.Handler {
	t.Helper()

	// Find templates directory from expected paths
	candidates := []string{
		"pkg/tally/templates",
		"../../pkg/tally/templates",
		"../../../pkg/tally/templates",
	}
	var found string
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			found = candidate
			break
		}
	}
	if found == "" {
		t.Fatal("Could not find pkg/tally/templates from any expected path")
	}

	os.Setenv("TALLY_TEMPLATES_DIR", found)
	t.Cleanup(func() { os.Unsetenv("TALLY_TEMPLATES_DIR") })

	host := os.Getenv("TALLY_HOST")
	if host == "" {
		t.Skip("TALLY_HOST not set — skipping integration test")
	}
	port := 9900
	if v := os.Getenv("TALLY_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}
	company := os.Getenv("TALLY_COMPANY")

	log, _ := logger.New("warn", "")
	client := tally.NewClient(host, port, 30)
	client.SetCompany(company)

	if err := client.Ping(); err != nil {
		t.Fatalf("cannot connect to Tally at %s:%d: %v", host, port, err)
	}

	registry, err := tally.LoadRegistry("pkg/tally/templates")
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	return mcp.NewHandler(client, registry, log)
}

func TestGetCompaniesIntegration(t *testing.T) {
	handler := setupHandler(t)
	result, err := handler.HandleToolCall("get_companies", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_companies failed: %v", err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Fatal("expected success=true")
	}
	companies := m["companies"].([]map[string]interface{})
	if len(companies) == 0 {
		t.Fatal("expected at least one company")
	}
	for i, c := range companies {
		if c["name"] == "" || c["name"] == nil {
			t.Errorf("company %d has empty name", i)
		}
		t.Logf("  Company %d: name=%v guid=%v", i+1, c["name"], c["guid"])
	}
	t.Logf("✓ get_companies: %d companies", len(companies))
}

func TestGetLedgersIntegration(t *testing.T) {
	handler := setupHandler(t)
	result, err := handler.HandleToolCall("get_ledgers", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_ledgers failed: %v", err)
	}
	m := result.(map[string]interface{})
	ledgers := m["ledgers"].([]map[string]interface{})
	t.Logf("✓ get_ledgers: %d ledgers", len(ledgers))
	for i, l := range ledgers {
		if i >= 3 {
			break
		}
		t.Logf("  Ledger %d: name=%v parent=%v", i+1, l["name"], l["parent"])
	}
}

func TestGetDebtorsIntegration(t *testing.T) {
	handler := setupHandler(t)
	result, err := handler.HandleToolCall("get_debtors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_debtors failed: %v", err)
	}
	m := result.(map[string]interface{})
	debtors := m["debtors"].([]map[string]interface{})
	t.Logf("✓ get_debtors: %d debtors", len(debtors))
	for i, d := range debtors {
		if i >= 3 {
			break
		}
		t.Logf("  Debtor %d: name=%v balance=%v", i+1, d["name"], d["closing_balance"])
	}
}

func TestGetCreditorsIntegration(t *testing.T) {
	handler := setupHandler(t)
	result, err := handler.HandleToolCall("get_creditors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_creditors failed: %v", err)
	}
	m := result.(map[string]interface{})
	creditors := m["creditors"].([]map[string]interface{})
	t.Logf("✓ get_creditors: %d creditors", len(creditors))
	for i, c := range creditors {
		if i >= 3 {
			break
		}
		t.Logf("  Creditor %d: name=%v balance=%v", i+1, c["name"], c["closing_balance"])
	}
}

func TestRegistryHasAllExpectedTools(t *testing.T) {
	// Set up templates directory
	candidates := []string{
		"pkg/tally/templates",
		"../../pkg/tally/templates",
		"../../../pkg/tally/templates",
	}
	var found string
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			found = candidate
			break
		}
	}
	if found == "" {
		t.Fatal("Could not find pkg/tally/templates from any expected path")
	}

	os.Setenv("TALLY_TEMPLATES_DIR", found)
	t.Cleanup(func() { os.Unsetenv("TALLY_TEMPLATES_DIR") })

	registry, err := tally.LoadRegistry(found)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	expected := []string{
		"get_companies", "get_ledgers", "get_ledger_details",
		"create_ledger", "get_debtors", "get_creditors", "get_creditor_vouchers",
	}
	for _, name := range expected {
		if registry.Get(name) == nil {
			t.Errorf("tool %q not found in registry", name)
		}
	}
	t.Logf("✓ registry has %d tools", len(registry.All()))
}

func TestAllGetToolsSequenceIntegration(t *testing.T) {
	handler := setupHandler(t)

	tools := []struct {
		name   string
		params map[string]interface{}
		key    string
	}{
		{"get_companies", nil, "companies"},
		{"get_ledgers", nil, "ledgers"},
		{"get_debtors", nil, "debtors"},
		{"get_creditors", nil, "creditors"},
	}

	for _, tc := range tools {
		result, err := handler.HandleToolCall(tc.name, tc.params)
		if err != nil {
			t.Fatalf("%s failed: %v", tc.name, err)
		}
		m := result.(map[string]interface{})
		items := m[tc.key].([]map[string]interface{})
		t.Logf("✓ %s: %d items", tc.name, len(items))
	}
}

func TestGetCreditorVouchersIntegration(t *testing.T) {
	handler := setupHandler(t)

	// First get creditors to find one to test with
	credResult, err := handler.HandleToolCall("get_creditors", map[string]interface{}{})
	if err != nil {
		t.Fatalf("get_creditors failed: %v", err)
	}
	credMap := credResult.(map[string]interface{})
	creditors := credMap["creditors"].([]map[string]interface{})
	if len(creditors) == 0 {
		t.Skip("No creditors found in Tally — skipping test")
	}

	// Use the first creditor
	creditorName := creditors[0]["name"].(string)
	t.Logf("Testing with creditor: %s", creditorName)

	// Test get_creditor_vouchers with date range
	// Using a small date range to avoid memory issues
	params := map[string]interface{}{
		"creditor_ledger_name": creditorName,
		"start_date":           "20260401",
		"end_date":             "20260430",
	}

	result, err := handler.HandleToolCall("get_creditor_vouchers", params)
	if err != nil {
		t.Fatalf("get_creditor_vouchers failed: %v", err)
	}

	m := result.(map[string]interface{})
	vouchers, ok := m["vouchers"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected vouchers array in response, got %T", m["vouchers"])
	}

	t.Logf("✓ get_creditor_vouchers for %q: %d vouchers", creditorName, len(vouchers))
	for i, v := range vouchers {
		if i >= 3 {
			break
		}
		t.Logf("  Voucher %d: number=%v date=%v reference=%v", i+1, v["voucher_number"], v["date"], v["reference"])
	}
}
