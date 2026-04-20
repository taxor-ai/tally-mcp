//go:build integration

package main

import (
	"testing"
	"time"
)

// TestCreateJournalVoucher creates a journal voucher for Cursor using the same
// ledgers as existing vouchers, then verifies it appears in get_creditor_vouchers.
func TestCreateJournalVoucher(t *testing.T) {
	handler := setupHandler(t)

	// Step 1: Get current voucher count for Cursor
	before, err := handler.HandleToolCall("get_creditor_vouchers", map[string]interface{}{
		"creditor_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_creditor_vouchers (before) failed: %v", err)
	}
	beforeVouchers := before.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("Cursor vouchers before: %d", len(beforeVouchers))

	// Step 2: Create a journal voucher using the same ledgers as existing vouchers:
	//   Subscription Charges + IGST Input RCM (debit) → IGST RCM Payable A/c + Cursor (credit)
	result, err := handler.HandleToolCall("create_journal_voucher", map[string]interface{}{
		"date":      "20260401",
		"reference": "TEST-JNL-APR-01",
		"narration": "Cursor Pro Apr 1 – May 1, 2026",
		"lines": []map[string]interface{}{
			{"ledger_name": "Subscription Charges", "entry_type": "debit", "amount": 1722.96},
			{"ledger_name": "IGST Input RCM", "entry_type": "debit", "amount": 310.0},
			{"ledger_name": "IGST RCM Payable A/c", "entry_type": "credit", "amount": 310.0},
			{"ledger_name": "Cursor", "entry_type": "credit", "amount": 1722.96},
		},
	})
	if err != nil {
		t.Fatalf("create_journal_voucher failed: %v", err)
	}

	// Step 3: Verify by XML response (created=1)
	m := result.(map[string]interface{})
	t.Logf("Create response: success=%v created=%v altered=%v error=%v", m["success"], m["created"], m["altered"], m["error"])
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_journal_voucher: success=false, error=%v", m["error"])
	}
	if created, _ := m["created"].(int); created != 1 {
		t.Errorf("expected created=1, got %v", m["created"])
	}
	t.Logf("✓ Journal voucher created (created=%v)", m["created"])

	// Step 4: Verify by calling get_creditor_vouchers — count must increase by 1
	time.Sleep(300 * time.Millisecond)
	after, err := handler.HandleToolCall("get_creditor_vouchers", map[string]interface{}{
		"creditor_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_creditor_vouchers (after) failed: %v", err)
	}
	afterVouchers := after.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("Cursor vouchers after: %d", len(afterVouchers))

	if len(afterVouchers) != len(beforeVouchers)+1 {
		t.Fatalf("expected %d vouchers after creation, got %d", len(beforeVouchers)+1, len(afterVouchers))
	}

	// Find and log the new voucher
	for _, v := range afterVouchers {
		if v["reference"] == "TEST-JNL-APR-01" {
			t.Logf("✓ New voucher confirmed: number=%v date=%v reference=%v", v["voucher_number"], v["date"], v["reference"])
			return
		}
	}
	t.Errorf("new voucher with reference TEST-JNL-APR-01 not found in results")
}
