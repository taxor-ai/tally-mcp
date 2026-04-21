package tally

import (
	"testing"
	"strings"
)

func TestLoadTemplate(t *testing.T) {
	params := map[string]interface{}{
		"ledger_name":    "Office Supplies",
		"parent":         "Indirect Expenses",
		"company_name":   "Test Company",
	}

	xml, err := LoadTemplate("ledger/create_ledger", params)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}

	if !strings.Contains(xml, "Office Supplies") {
		t.Error("Template not properly parameterized: missing ledger name")
	}
}

func TestLoadTemplateWithoutParams(t *testing.T) {
	xml, err := LoadTemplate("ledger/get_ledgers", map[string]interface{}{})
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}

	if xml == "" {
		t.Error("Template should not be empty")
	}
	if !strings.Contains(xml, "<ENVELOPE>") {
		t.Error("Template should contain ENVELOPE element")
	}
}
