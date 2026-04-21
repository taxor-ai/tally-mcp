//go:build integration

package main

import (
	"testing"
)

func TestGetSalesVouchers(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_sales_vouchers", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_sales_vouchers failed: %v", err)
	}

	m := result.(map[string]interface{})
	vouchers, ok := m["vouchers"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected vouchers array in response, got %T", m["vouchers"])
	}

	if len(vouchers) == 0 {
		t.Fatal("expected at least one Sales voucher for TestStore, got none")
	}

	for i, v := range vouchers {
		voucherType, _ := v["voucher_type"].(string)
		if voucherType != "Sales" {
			t.Errorf("voucher %d: expected type=Sales, got %q", i+1, voucherType)
		}
		t.Logf("  Voucher %d: number=%v date=%v type=%v narration=%v", i+1, v["voucher_number"], v["date"], voucherType, v["narration"])
		if entries, ok := v["ledger_entries"].([]map[string]interface{}); ok {
			for _, e := range entries {
				t.Logf("    ledger=%v amount=%v", e["ledger_name"], e["amount"])
			}
		}
	}
	t.Logf("✓ get_sales_vouchers (TestStore): %d Sales vouchers", len(vouchers))
}
