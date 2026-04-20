//go:build integration

package main

import (
	"testing"
)

func TestGetDebtorVouchers(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("get_debtor_vouchers", map[string]interface{}{
		"debtor_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_debtor_vouchers failed: %v", err)
	}

	m := result.(map[string]interface{})
	vouchers, ok := m["vouchers"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected vouchers array in response, got %T", m["vouchers"])
	}

	if len(vouchers) != 2 {
		t.Fatalf("expected exactly 2 vouchers for TestStore, got %d", len(vouchers))
	}

	for i, v := range vouchers {
		voucherType, _ := v["voucher_type"].(string)
		if voucherType != "Sales" {
			t.Errorf("voucher %d: expected type=Sales, got %q", i+1, voucherType)
		}
		t.Logf("  Voucher %d: number=%v date=%v type=%v", i+1, v["voucher_number"], v["date"], voucherType)
	}
	t.Logf("✓ get_debtor_vouchers (TestStore): %d Sales vouchers", len(vouchers))
}
