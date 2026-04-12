package tally

import (
	"testing"
)

func TestLedgerModel(t *testing.T) {
	ledger := Ledger{
		Name:        "Office Supplies",
		Type:        "Expense",
		Balance:     1500.00,
		Description: "Office supplies expenses",
	}

	if ledger.Name != "Office Supplies" {
		t.Errorf("Expected name 'Office Supplies', got '%s'", ledger.Name)
	}
	if ledger.Type != "Expense" {
		t.Errorf("Expected type 'Expense', got '%s'", ledger.Type)
	}
}

func TestVoucherModel(t *testing.T) {
	voucher := Voucher{
		VoucherID:  "INV-001",
		Type:       "Invoice",
		Date:       "2026-04-10",
		Party:      "ABC Corp",
		Amount:     5000.00,
		Status:     "Open",
	}

	if voucher.VoucherID != "INV-001" {
		t.Errorf("Expected ID 'INV-001', got '%s'", voucher.VoucherID)
	}
}
