//go:build integration

package main

import (
	"testing"
)

func TestGetCreditorVouchers(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_creditor_vouchers", map[string]interface{}{
		"creditor_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_creditor_vouchers failed: %v", err)
	}

	m := result.(map[string]interface{})
	vouchers, ok := m["vouchers"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected vouchers array in response, got %T", m["vouchers"])
	}

	if len(vouchers) == 0 {
		t.Fatal("expected at least one voucher for Cursor")
	}

	for i, v := range vouchers {
		voucherType, _ := v["voucher_type"].(string)
		if voucherType != "Journal" {
			t.Errorf("voucher %d: expected type=Journal, got %q", i+1, voucherType)
		}
	}
	t.Logf("✓ get_creditor_vouchers (Cursor): %d Journal vouchers", len(vouchers))
	for i, v := range vouchers {
		if i >= 3 {
			break
		}
		t.Logf("  Voucher %d: number=%v date=%v reference=%v", i+1, v["voucher_number"], v["date"], v["reference"])
	}
}
