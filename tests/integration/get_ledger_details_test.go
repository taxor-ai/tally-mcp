//go:build integration

package main

import (
	"testing"
)

func TestGetLedgerDetails(t *testing.T) {
	handler := setupHandler(t)

	// Get the list of ledgers and pick the first one to query details for
	ledgersResult, err := handler.HandleToolCall("get_ledgers", nil)
	if err != nil {
		t.Fatalf("get_ledgers failed: %v", err)
	}
	ledgersMap := ledgersResult.(map[string]interface{})
	ledgers := ledgersMap["ledgers"].([]map[string]interface{})
	if len(ledgers) == 0 {
		t.Skip("no ledgers in Tally — skipping get_ledger_details test")
	}

	ledgerName := ledgers[0]["name"].(string)
	t.Logf("Fetching details for ledger: %s", ledgerName)

	result, err := handler.HandleToolCall("get_ledger_details", map[string]interface{}{
		"ledger_name": ledgerName,
	})
	if err != nil {
		t.Fatalf("get_ledger_details failed: %v", err)
	}

	m := result.(map[string]interface{})
	ledger, ok := m["ledger"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected ledger object in response, got: %v", m)
	}
	if ledger["name"] == nil || ledger["name"] == "" {
		t.Fatalf("expected ledger name in response, got: %v", ledger)
	}
	if ledger["name"].(string) != ledgerName {
		t.Errorf("expected name=%q, got %q", ledgerName, ledger["name"])
	}

	t.Logf("✓ get_ledger_details: name=%v parent=%v", ledger["name"], ledger["parent"])
}
