//go:build integration

package main

import (
	"testing"
	"time"
)

// TestCreatePayment creates a payment voucher for a creditor using the same
// ledgers as existing vouchers, then verifies it appears in get_payments.
func TestCreatePayment(t *testing.T) {
	handler := setupHandler(t)

	// Step 1: Get current payment count for Cursor
	before, err := handler.HandleToolCall("get_payments", map[string]interface{}{
		"party_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_payments (before) failed: %v", err)
	}
	beforeVouchers := before.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("Cursor payments before: %d", len(beforeVouchers))

	// Step 2: Create a payment voucher
	// Payment structure: creditor ledger (credit, positive amount) + bank account (debit, negative amount)
	result, err := handler.HandleToolCall("create_payment", map[string]interface{}{
		"date":              "20260401",
		"reference":         "TEST-PMT-001",
		"narration":         "Test payment for April purchase",
		"party_ledger_name": "Cursor",
		"lines": []map[string]interface{}{
			{"ledger_name": "Cursor", "amount": 1000},             // Credit (payment to vendor)
			{"ledger_name": "ICICI Bank", "amount": -1000}, // Debit (from bank)
		},
	})
	if err != nil {
		t.Fatalf("create_payment failed: %v", err)
	}

	// Step 3: Verify by XML response
	m := result.(map[string]interface{})
	t.Logf("Create response: success=%v created=%v altered=%v error=%v", m["success"], m["created"], m["altered"], m["error"])
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_payment: success=false, error=%v", m["error"])
	}
	if created, _ := m["created"].(int); created != 1 {
		t.Errorf("expected created=1, got %v", m["created"])
	}
	t.Logf("✓ Payment voucher created (created=%v)", m["created"])

	// Step 4: Verify by calling get_payments — count must increase by 1
	time.Sleep(300 * time.Millisecond)
	after, err := handler.HandleToolCall("get_payments", map[string]interface{}{
		"party_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_payments (after) failed: %v", err)
	}
	afterVouchers := after.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("Cursor payments after: %d", len(afterVouchers))

	if len(afterVouchers) != len(beforeVouchers)+1 {
		t.Fatalf("expected %d vouchers after creation, got %d", len(beforeVouchers)+1, len(afterVouchers))
	}

	// Test passed successfully
	t.Logf("✓ New payment voucher successfully created and verified")
}

// TestCreatePaymentWithBillAllocations creates a payment and allocates it against bills
// using different Tally allocation types: "Agnst Ref", "New Ref", "Advance", "On Account"
func TestCreatePaymentWithBillAllocations(t *testing.T) {
	handler := setupHandler(t)

	// Create a payment with "Agnst Ref" allocation (against an existing invoice/purchase order)
	result, err := handler.HandleToolCall("create_payment", map[string]interface{}{
		"date":              "20260401",
		"reference":         "CHQ-002",
		"narration":         "Payment against Invoice INV-2026-001",
		"party_ledger_name": "Cursor",
		"lines": []map[string]interface{}{
			{
				"ledger_name": "Cursor",
				"amount":      5000,
				"bill_allocations": []map[string]interface{}{
					{
						"name":   "INV-2026-001",
						"type":   "Agnst Ref",
						"amount": 5000,
					},
				},
			},
			{"ledger_name": "ICICI Bank", "amount": -5000},
		},
	})
	if err != nil {
		t.Fatalf("create_payment with bill allocations failed: %v", err)
	}

	m := result.(map[string]interface{})
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_payment with allocations: success=false, error=%v", m["error"])
	}
	t.Logf("✓ Payment voucher with 'Agnst Ref' allocation created (created=%v)", m["created"])
}

// TestCreatePaymentWithAdvanceAllocation creates a payment marked as advance
func TestCreatePaymentWithAdvanceAllocation(t *testing.T) {
	handler := setupHandler(t)

	// Create a payment marked as "Advance" (advance payment for future purchases)
	result, err := handler.HandleToolCall("create_payment", map[string]interface{}{
		"date":              "20260401",
		"reference":         "CHQ-003",
		"narration":         "Advance payment for future purchases",
		"party_ledger_name": "Cursor",
		"lines": []map[string]interface{}{
			{
				"ledger_name": "Cursor",
				"amount":      10000,
				"bill_allocations": []map[string]interface{}{
					{
						"name":   "ADV-2026-001",
						"type":   "Advance",
						"amount": 10000,
					},
				},
			},
			{"ledger_name": "ICICI Bank", "amount": -10000},
		},
	})
	if err != nil {
		t.Fatalf("create_payment with advance allocation failed: %v", err)
	}

	m := result.(map[string]interface{})
	if success, _ := m["success"].(bool); !success {
		t.Fatalf("create_payment advance: success=false, error=%v", m["error"])
	}
	t.Logf("✓ Payment voucher with 'Advance' allocation created (created=%v)", m["created"])
}
