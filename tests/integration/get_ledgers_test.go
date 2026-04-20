//go:build integration

package main

import (
	"testing"
)

func TestGetLedgers(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_ledgers", nil)
	if err != nil {
		t.Fatalf("get_ledgers failed: %v", err)
	}

	m := result.(map[string]interface{})
	ledgers := m["ledgers"].([]map[string]interface{})
	if len(ledgers) == 0 {
		t.Fatal("expected at least one ledger")
	}

	for i, l := range ledgers {
		if l["name"] == "" || l["name"] == nil {
			t.Errorf("ledger %d has empty name", i+1)
		}
	}
	t.Logf("✓ get_ledgers: %d ledgers", len(ledgers))
	for i, l := range ledgers {
		if i >= 5 {
			break
		}
		t.Logf("  Ledger %d: name=%v parent=%v", i+1, l["name"], l["parent"])
	}
}
