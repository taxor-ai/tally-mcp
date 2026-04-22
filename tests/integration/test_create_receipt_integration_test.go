//go:build integration

package main

import (
	"testing"
	"time"
)

// TestCreateReceipt creates a receipt voucher for TestStore (debtor), then verifies
// it appears in get_receipts. Sign convention mirrors existing Tally receipts:
// - Party (TestStore): positive amount (credit — receivable reduced)
// - Bank (ICICI Bank): negative amount (debit — money received into bank)
func TestCreateReceipt(t *testing.T) {
	handler := setupHandler(t)

	// Step 1: Get current receipt count for TestStore
	before, err := handler.HandleToolCall("get_receipts", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_receipts (before) failed: %v", err)
	}
	beforeVouchers := before.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("TestStore receipts before: %d", len(beforeVouchers))

	// Step 2: Create a receipt voucher
	// Party (TestStore) comes FIRST so Tally shows it in the Particulars column.
	// Party: +6800 (credit, receivable reduced), Bank: -6800 (debit, money received).
	// "New Ref" creates a fresh bill reference — no prior outstanding invoice required.
	result, err := handler.HandleToolCall("create_receipt", map[string]interface{}{
		"date":              "20260401",
		"reference":         "NEFT-REC-001",
		"narration":         "NEFT-UTIBN62026040112345678-TEST STORE-/////-912020050131310-UTIB0000333",
		"party_ledger_name": "TestStore",
		"bank_account":      "ICICI Bank",
		"lines": []map[string]interface{}{
			{
				"ledger_name": "TestStore",
				"amount":      6800,
				"is_party":    true,
				"bill_allocations": []map[string]interface{}{
					{
						"name":   "INV2526000001",
						"type":   "New Ref",
						"amount": 6800,
					},
				},
			},
			{"ledger_name": "ICICI Bank", "amount": -6800, "is_party": false},
		},
	})
	if err != nil {
		t.Fatalf("create_receipt failed: %v", err)
	}

	// Step 3: Verify via XML response
	m := result.(map[string]interface{})
	t.Logf("Create response: success=%v created=%v altered=%v error=%v", m["success"], m["created"], m["altered"], m["error"])
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_receipt: success=false, error=%v", m["error"])
	}
	if created, _ := m["created"].(int); created != 1 {
		t.Errorf("expected created=1, got %v", m["created"])
	}
	t.Logf("✓ Receipt voucher created (created=%v)", m["created"])

	// Step 4: Verify by calling get_receipts — count must increase by 1
	time.Sleep(300 * time.Millisecond)
	after, err := handler.HandleToolCall("get_receipts", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_receipts (after) failed: %v", err)
	}
	afterVouchers := after.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("TestStore receipts after: %d", len(afterVouchers))

	if len(afterVouchers) != len(beforeVouchers)+1 {
		t.Fatalf("expected %d vouchers after creation, got %d", len(beforeVouchers)+1, len(afterVouchers))
	}

	t.Logf("✓ New receipt voucher successfully created and verified")
}

// TestCreateReceiptWithBillAllocations creates a receipt and allocates it against a new
// bill reference — use "New Ref" to create a reference for future reconciliation.
func TestCreateReceiptWithBillAllocations(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("create_receipt", map[string]interface{}{
		"date":              "20260401",
		"reference":         "NEFT-REC-002",
		"narration":         "NEFT-UTIBN62026040198765432-TEST STORE-/////-912020050131310-UTIB0000333",
		"party_ledger_name": "TestStore",
		"bank_account":      "ICICI Bank",
		"lines": []map[string]interface{}{
			{
				"ledger_name": "TestStore",
				"amount":      6800,
				"is_party":    true,
				"bill_allocations": []map[string]interface{}{
					{
						"name":   "INV2526000002",
						"type":   "New Ref",
						"amount": 6800,
					},
				},
			},
			{"ledger_name": "ICICI Bank", "amount": -6800, "is_party": false},
		},
	})
	if err != nil {
		t.Fatalf("create_receipt with bill allocations failed: %v", err)
	}

	m := result.(map[string]interface{})
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_receipt with allocations: success=false, error=%v", m["error"])
	}
	t.Logf("✓ Receipt voucher with 'New Ref' allocation created (created=%v)", m["created"])
}

// TestCreateReceiptWithAdvanceAllocation creates a receipt marked as advance
// (money received before an invoice is raised).
func TestCreateReceiptWithAdvanceAllocation(t *testing.T) {
	handler := setupHandler(t)

	result, err := handler.HandleToolCall("create_receipt", map[string]interface{}{
		"date":              "20260401",
		"reference":         "NEFT-REC-003",
		"narration":         "Advance receipt from TestStore",
		"party_ledger_name": "TestStore",
		"bank_account":      "ICICI Bank",
		"lines": []map[string]interface{}{
			{
				"ledger_name": "TestStore",
				"amount":      10000,
				"is_party":    true,
				"bill_allocations": []map[string]interface{}{
					{
						"name":   "ADV-REC-2026-001",
						"type":   "Advance",
						"amount": 10000,
					},
				},
			},
			{"ledger_name": "ICICI Bank", "amount": -10000, "is_party": false},
		},
	})
	if err != nil {
		t.Fatalf("create_receipt with advance allocation failed: %v", err)
	}

	m := result.(map[string]interface{})
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_receipt advance: success=false, error=%v", m["error"])
	}
	t.Logf("✓ Receipt voucher with 'Advance' allocation created (created=%v)", m["created"])
}
