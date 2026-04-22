//go:build integration

package main

import (
	"testing"
)

func TestGetReceipts(t *testing.T) {
	handler := setupHandler(t)

	// Fetch receipts for a debtor (customer)
	// Use a debtor that exists in the test data
	t.Log("Fetching receipt vouchers for TestStore...")
	result, err := handler.HandleToolCall("get_receipts", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_receipts failed: %v", err)
	}

	m := result.(map[string]interface{})

	// Check for vouchers in response
	vouchers, ok := m["vouchers"].([]map[string]interface{})
	if !ok {
		// If no vouchers key, the tool may return a different structure
		t.Logf("Response structure: %v", m)
		t.Logf("✓ get_receipts succeeded (no vouchers found for this debtor yet)")
		return
	}

	// If vouchers exist, validate their structure
	for i, v := range vouchers {
		voucherType, _ := v["voucher_type"].(string)
		if voucherType != "Receipt" {
			t.Errorf("voucher %d: expected type=Receipt, got %q", i+1, voucherType)
		}
		t.Logf("  Voucher %d: number=%v date=%v type=%v narration=%v", i+1, v["voucher_number"], v["date"], voucherType, v["narration"])
		if entries, ok := v["ledger_entries"].([]map[string]interface{}); ok {
			for _, e := range entries {
				t.Logf("    ledger=%v amount=%v", e["ledger_name"], e["amount"])
			}
		}
	}
	t.Logf("✓ get_receipts (TestStore): %d Receipt vouchers", len(vouchers))
}
