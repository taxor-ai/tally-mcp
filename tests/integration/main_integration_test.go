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

// setupHandler creates an MCP handler for integration tests.
// It skips the test if TALLY_HOST is not set or Tally is unreachable.
func setupHandler(t *testing.T) *mcp.Handler {
	t.Helper()

	candidates := []string{
		"pkg/tally/templates",
		"../../pkg/tally/templates",
	}
	var templatesDir string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			templatesDir = c
			break
		}
	}
	if templatesDir == "" {
		t.Fatal("could not find pkg/tally/templates")
	}
	os.Setenv("TALLY_TEMPLATES_DIR", templatesDir)
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

	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	return mcp.NewHandler(client, registry, log)
}

func TestRegistryHasAllExpectedTools(t *testing.T) {
	candidates := []string{
		"pkg/tally/templates",
		"../../pkg/tally/templates",
	}
	var templatesDir string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			templatesDir = c
			break
		}
	}
	if templatesDir == "" {
		t.Fatal("could not find pkg/tally/templates")
	}
	os.Setenv("TALLY_TEMPLATES_DIR", templatesDir)
	t.Cleanup(func() { os.Unsetenv("TALLY_TEMPLATES_DIR") })

	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}

	expected := []string{
		"get_companies",
		"get_ledgers",
		"get_ledger_details",
		"create_ledger",
		"get_debtors",
		"get_creditors",
		"get_debtor_vouchers",
		"get_creditor_vouchers",
		"create_journal_voucher",
		"create_sales_voucher",
	}
	for _, name := range expected {
		if registry.Get(name) == nil {
			t.Errorf("tool %q not found in registry", name)
		}
	}
	t.Logf("✓ registry has %d tools", len(registry.All()))
}
