//go:build integration

package main

import (
	"testing"
	"time"
)

// TestCreateSalesVoucher creates a sales voucher for TestStore using the same
// ledgers as existing vouchers, then verifies it appears in get_sales_vouchers.
func TestCreateSalesVoucher(t *testing.T) {
	handler := setupHandler(t)

	// Step 1: Get current voucher count for TestStore
	before, err := handler.HandleToolCall("get_sales_vouchers", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_sales_vouchers (before) failed: %v", err)
	}
	beforeVouchers := before.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("TestStore vouchers before: %d", len(beforeVouchers))

	// Step 2: Create a sales voucher using the same ledgers as existing vouchers:
	//   TestStore (debit) → Delivery Management Software Service-GST + Output SGST + Output CGST (credit)
	result, err := handler.HandleToolCall("create_sales_voucher", map[string]interface{}{
		"date":               "20260402",
		"reference":          "TEST-SAL-APR-02",
		"narration":          "Being the sale for the month of April'26",
		"party_ledger_name": "TestStore",
		"lines": []map[string]interface{}{
			{"ledger_name": "TestStore", "entry_type": "debit", "amount": 2.36},
			{"ledger_name": "Delivery Management Software Service-GST", "entry_type": "credit", "amount": 2.0},
			{"ledger_name": "Output SGST", "entry_type": "credit", "amount": 0.18},
			{"ledger_name": "Output CGST", "entry_type": "credit", "amount": 0.18},
		},
	})
	if err != nil {
		t.Fatalf("create_sales_voucher failed: %v", err)
	}

	// Step 3: Verify by XML response (created=1)
	m := result.(map[string]interface{})
	t.Logf("Create response: success=%v created=%v altered=%v error=%v", m["success"], m["created"], m["altered"], m["error"])
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_sales_voucher: success=false, error=%v", m["error"])
	}
	if created, _ := m["created"].(int); created != 1 {
		t.Errorf("expected created=1, got %v", m["created"])
	}
	t.Logf("✓ Sales voucher created (created=%v)", m["created"])

	// Step 4: Verify by calling get_sales_vouchers — count must increase by 1
	time.Sleep(300 * time.Millisecond)
	after, err := handler.HandleToolCall("get_sales_vouchers", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_sales_vouchers (after) failed: %v", err)
	}
	afterVouchers := after.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("TestStore vouchers after: %d", len(afterVouchers))

	if len(afterVouchers) != len(beforeVouchers)+1 {
		t.Fatalf("expected %d vouchers after creation, got %d", len(beforeVouchers)+1, len(afterVouchers))
	}

	// Find and log the new voucher
	for _, v := range afterVouchers {
		if v["reference"] == "TEST-SAL-APR-02" {
			t.Logf("✓ New voucher confirmed: number=%v date=%v reference=%v", v["voucher_number"], v["date"], v["reference"])
			return
		}
	}
	t.Errorf("new voucher with reference TEST-SAL-APR-01 not found in results")
}
