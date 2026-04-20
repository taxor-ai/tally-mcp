//go:build integration

package main

import (
	"testing"
)

func TestGetDebtors(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_debtors", nil)
	if err != nil {
		t.Fatalf("get_debtors failed: %v", err)
	}

	m := result.(map[string]interface{})
	debtors := m["debtors"].([]map[string]interface{})
	if len(debtors) == 0 {
		t.Fatal("expected at least one debtor")
	}

	for i, d := range debtors {
		if d["name"] == "" || d["name"] == nil {
			t.Errorf("debtor %d has empty name", i+1)
		}
	}
	t.Logf("✓ get_debtors: %d debtors", len(debtors))
	for i, d := range debtors {
		if i >= 5 {
			break
		}
		t.Logf("  Debtor %d: name=%v closing_balance=%v", i+1, d["name"], d["closing_balance"])
	}
}
