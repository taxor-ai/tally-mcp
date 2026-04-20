//go:build integration

package main

import (
	"testing"
)

func TestGetCompanies(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_companies", nil)
	if err != nil {
		t.Fatalf("get_companies failed: %v", err)
	}

	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Fatalf("expected success=true, got %v", m["success"])
	}

	companies := m["companies"].([]map[string]interface{})
	if len(companies) == 0 {
		t.Fatal("expected at least one company")
	}

	for i, c := range companies {
		if c["name"] == "" || c["name"] == nil {
			t.Errorf("company %d has empty name", i+1)
		}
		t.Logf("  Company %d: name=%v guid=%v", i+1, c["name"], c["guid"])
	}
	t.Logf("✓ get_companies: %d companies", len(companies))
}
