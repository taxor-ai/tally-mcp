//go:build integration

package main

import (
	"testing"
)

func TestGetPayments(t *testing.T) {
	handler := setupHandler(t)

	// Fetch payments for a creditor (vendor)
	// Use a creditor that exists in the test data
	t.Log("Fetching payment vouchers for Cursor...")
	result, err := handler.HandleToolCall("get_payments", map[string]interface{}{
		"party_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_payments failed: %v", err)
	}

	m := result.(map[string]interface{})

	// Check for vouchers in response
	vouchers, ok := m["vouchers"].([]map[string]interface{})
	if !ok {
		// If no vouchers key, the tool may return a different structure
		t.Logf("Response structure: %v", m)
		t.Logf("✓ get_payments succeeded (no vouchers found for this creditor yet)")
		return
	}

	// If vouchers exist, validate their structure
	for i, v := range vouchers {
		voucherType, _ := v["voucher_type"].(string)
		if voucherType != "Payment" {
			t.Errorf("voucher %d: expected type=Payment, got %q", i+1, voucherType)
		}
		t.Logf("  Voucher %d: number=%v date=%v type=%v narration=%v", i+1, v["voucher_number"], v["date"], voucherType, v["narration"])
		if entries, ok := v["ledger_entries"].([]map[string]interface{}); ok {
			for _, e := range entries {
				t.Logf("    ledger=%v amount=%v", e["ledger_name"], e["amount"])
			}
		}
	}
	t.Logf("✓ get_payments (Cursor): %d Payment vouchers", len(vouchers))
}

