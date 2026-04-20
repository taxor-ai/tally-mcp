//go:build integration

package main

import (
	"testing"
)

func TestGetCreditors(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_creditors", nil)
	if err != nil {
		t.Fatalf("get_creditors failed: %v", err)
	}

	m := result.(map[string]interface{})
	creditors := m["creditors"].([]map[string]interface{})
	if len(creditors) == 0 {
		t.Fatal("expected at least one creditor")
	}

	for i, c := range creditors {
		if c["name"] == "" || c["name"] == nil {
			t.Errorf("creditor %d has empty name", i+1)
		}
	}
	t.Logf("✓ get_creditors: %d creditors", len(creditors))
	for i, c := range creditors {
		if i >= 5 {
			break
		}
		t.Logf("  Creditor %d: name=%v closing_balance=%v", i+1, c["name"], c["closing_balance"])
	}
}
