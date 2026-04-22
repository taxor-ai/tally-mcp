//go:build integration

package main

import (
	"testing"
	"time"
)

// TestSalesAndPaymentWorkflow tests the complete workflow:
// 1. Create a sales voucher for TestStore (a customer/debtor)
// 2. Create a payment voucher for Cursor (a vendor/creditor)
// 3. Verify both are created via get methods
func TestSalesAndPaymentWorkflow(t *testing.T) {
	handler := setupHandler(t)

	// Step 1: Get initial sales vouchers count for TestStore
	t.Log("Step 1: Get initial sales vouchers for TestStore...")
	beforeSales, err := handler.HandleToolCall("get_sales_vouchers", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_sales_vouchers (before) failed: %v", err)
	}
	beforeSalesVouchers := beforeSales.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("  TestStore sales vouchers before: %d", len(beforeSalesVouchers))

	// Step 2: Create a sales voucher for TestStore (using validated ledgers from test_create_sales_voucher_integration_test.go)
	t.Log("Step 2: Create sales voucher for TestStore...")
	salesResult, err := handler.HandleToolCall("create_sales_voucher", map[string]interface{}{
		"date":               "20260401",
		"reference":         "INV-WF-001",
		"narration":         "Sales voucher for workflow test",
		"party_ledger_name": "TestStore",
		"lines": []map[string]interface{}{
			{"ledger_name": "TestStore", "amount": -2.36},
			{"ledger_name": "Delivery Management Software Service-GST", "amount": 2.0},
			{"ledger_name": "Output SGST", "amount": 0.18},
			{"ledger_name": "Output CGST", "amount": 0.18},
		},
	})
	if err != nil {
		t.Fatalf("create_sales_voucher failed: %v", err)
	}

	salesMap := salesResult.(map[string]interface{})
	t.Logf("  Create response: success=%v created=%v", salesMap["success"], salesMap["created"])
	if success, _ := salesMap["success"].(bool); !success {
		t.Fatalf("create_sales_voucher failed: success=false, error=%v", salesMap["error"])
	}
	if created, _ := salesMap["created"].(int); created != 1 {
		t.Errorf("expected created=1, got %v", created)
	}
	t.Log("  ✓ Sales voucher created")

	// Step 3: Verify sales voucher via get_sales_vouchers
	time.Sleep(300 * time.Millisecond)
	t.Log("Step 3: Verify sales voucher was created...")
	afterSales, err := handler.HandleToolCall("get_sales_vouchers", map[string]interface{}{
		"party_ledger_name": "TestStore",
	})
	if err != nil {
		t.Fatalf("get_sales_vouchers (after sales) failed: %v", err)
	}
	afterSalesVouchers := afterSales.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("  TestStore sales vouchers after: %d", len(afterSalesVouchers))

	if len(afterSalesVouchers) != len(beforeSalesVouchers)+1 {
		t.Fatalf("expected %d sales vouchers after creation, got %d", len(beforeSalesVouchers)+1, len(afterSalesVouchers))
	}
	t.Log("  ✓ Sales voucher verified via get method")

	// Step 4: Get initial payment vouchers count for Cursor
	t.Log("Step 4: Get initial payment vouchers for Cursor...")
	beforePayments, err := handler.HandleToolCall("get_payments", map[string]interface{}{
		"party_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_payments (before) failed: %v", err)
	}
	beforePaymentVouchers := beforePayments.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("  Cursor payment vouchers before: %d", len(beforePaymentVouchers))

	// Step 5: Create a payment voucher for Cursor (using validated ledgers from test_create_payment_integration_test.go)
	t.Log("Step 5: Create payment voucher for Cursor...")
	paymentResult, err := handler.HandleToolCall("create_payment", map[string]interface{}{
		"date":               "20260401",
		"reference":         "CHQ-WF-001",
		"narration":         "Payment voucher for workflow test",
		"party_ledger_name": "Cursor",
		"bank_account":      "ICICI Bank",
		"lines": []map[string]interface{}{
			{
				"ledger_name": "Cursor",
				"amount":      1500,
				"bill_allocations": []map[string]interface{}{
					{
						"name":   "INV-WF-001",
						"type":   "Agnst Ref",
						"amount": 1500,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create_payment failed: %v", err)
	}

	paymentMap := paymentResult.(map[string]interface{})
	t.Logf("  Create response: success=%v created=%v", paymentMap["success"], paymentMap["created"])
	if success, _ := paymentMap["success"].(bool); !success {
		t.Fatalf("create_payment failed: success=false, error=%v", paymentMap["error"])
	}
	if created, _ := paymentMap["created"].(int); created != 1 {
		t.Errorf("expected created=1, got %v", created)
	}
	t.Log("  ✓ Payment voucher created with bill allocation")

	// Step 6: Verify payment voucher via get_payments
	time.Sleep(300 * time.Millisecond)
	t.Log("Step 6: Verify payment voucher was created...")
	afterPayments, err := handler.HandleToolCall("get_payments", map[string]interface{}{
		"party_ledger_name": "Cursor",
	})
	if err != nil {
		t.Fatalf("get_payments (after) failed: %v", err)
	}
	afterPaymentVouchers := afterPayments.(map[string]interface{})["vouchers"].([]map[string]interface{})
	t.Logf("  Cursor payment vouchers after: %d", len(afterPaymentVouchers))

	if len(afterPaymentVouchers) != len(beforePaymentVouchers)+1 {
		t.Fatalf("expected %d payment vouchers after creation, got %d", len(beforePaymentVouchers)+1, len(afterPaymentVouchers))
	}
	t.Log("  ✓ Payment voucher verified via get method")

	// Step 7: Summary
	t.Log("")
	t.Log("✓ Complete workflow test passed:")
	t.Logf("  - Sales voucher created for TestStore (total: %d)", len(afterSalesVouchers))
	t.Logf("  - Payment voucher created for Cursor (total: %d)", len(afterPaymentVouchers))
	t.Logf("  - Both verified via get methods")
}
