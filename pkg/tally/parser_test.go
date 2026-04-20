package tally

import (
	"testing"
)

func TestNestedLedgerEntriesParsing(t *testing.T) {
	// Mock Tally XML with nested ledger entries
	xmlData := []byte(`<?xml version="1.0"?>
<RESPONSE>
  <COLLECTION>
    <VOUCHER>
      <VOUCHERNUMBER>415</VOUCHERNUMBER>
      <DATE>20250310</DATE>
      <REFERENCE></REFERENCE>
      <NARRATION>Being the sale for the month of March'25</NARRATION>
      <VOUCHERTYPENAME>Sales</VOUCHERTYPENAME>
      <ALLLEDGERENTRIES>
        <LIST>
          <ENTRY>
            <LEDGERNAME>TestStore</LEDGERNAME>
            <AMOUNT>-2.36</AMOUNT>
            <ISDEEMEDPOSITIVE>No</ISDEEMEDPOSITIVE>
            <DATE>20250310</DATE>
            <REFERENCE>INV2425000416</REFERENCE>
            <INVOICENUMBER>1525</INVOICENUMBER>
            <TAXAMOUNT>0</TAXAMOUNT>
          </ENTRY>
          <ENTRY>
            <LEDGERNAME>Sales Account</LEDGERNAME>
            <AMOUNT>2.36</AMOUNT>
            <ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
            <DATE>20250310</DATE>
            <REFERENCE></REFERENCE>
            <INVOICENUMBER></INVOICENUMBER>
            <TAXAMOUNT>0</TAXAMOUNT>
          </ENTRY>
        </LIST>
      </ALLLEDGERENTRIES>
    </VOUCHER>
  </COLLECTION>
</RESPONSE>`)

	spec := ParserSpec{
		Type:       "list",
		ItemsXPath: "//COLLECTION/VOUCHER",
		ResultKey:  "vouchers",
		Fields: map[string]FieldSpec{
			"voucher_number": {
				XPath: "VOUCHERNUMBER",
			},
			"date": {
				XPath:     "DATE",
				Transform: "tally_date",
			},
			"voucher_type": {
				XPath: "VOUCHERTYPENAME",
			},
			"ledger_entries": {
				ItemsXPath: "ALLLEDGERENTRIES/LIST/ENTRY",
				Fields: map[string]FieldSpec{
					"ledger_name": {
						XPath: "LEDGERNAME",
					},
					"amount": {
						XPath:     "AMOUNT",
						Transform: "number",
					},
					"debit_credit": {
						XPath:     "ISDEEMEDPOSITIVE",
						Transform: "boolean",
					},
					"date": {
						XPath:     "DATE",
						Transform: "tally_date",
					},
					"reference": {
						XPath: "REFERENCE",
					},
					"invoice_number": {
						XPath: "INVOICENUMBER",
					},
					"tax_amount": {
						XPath:     "TAXAMOUNT",
						Transform: "number",
					},
				},
			},
		},
	}

	result, err := ParseResponse(xmlData, spec)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if !result["success"].(bool) {
		t.Fatal("Parse was not successful")
	}

	vouchers := result["vouchers"].([]map[string]interface{})
	if len(vouchers) != 1 {
		t.Fatalf("Expected 1 voucher, got %d", len(vouchers))
	}

	voucher := vouchers[0]
	if voucher["voucher_number"] != "415" {
		t.Errorf("Wrong voucher number: %v", voucher["voucher_number"])
	}

	if voucher["date"] != "2025-03-10" {
		t.Errorf("Wrong date: %v", voucher["date"])
	}

	ledgerEntries := voucher["ledger_entries"].([]map[string]interface{})
	if len(ledgerEntries) != 2 {
		t.Fatalf("Expected 2 ledger entries, got %d", len(ledgerEntries))
	}

	// Check first entry
	entry1 := ledgerEntries[0]
	if entry1["ledger_name"] != "TestStore" {
		t.Errorf("Wrong ledger name: %v", entry1["ledger_name"])
	}
	if entry1["amount"] != -2.36 {
		t.Errorf("Wrong amount: %v", entry1["amount"])
	}
	if entry1["debit_credit"] != false {
		t.Errorf("Wrong debit_credit: %v", entry1["debit_credit"])
	}

	// Check second entry
	entry2 := ledgerEntries[1]
	if entry2["ledger_name"] != "Sales Account" {
		t.Errorf("Wrong ledger name: %v", entry2["ledger_name"])
	}
	if entry2["amount"] != 2.36 {
		t.Errorf("Wrong amount: %v", entry2["amount"])
	}
	if entry2["debit_credit"] != true {
		t.Errorf("Wrong debit_credit: %v", entry2["debit_credit"])
	}

	t.Logf("✓ Nested ledger entries parsed correctly!")
	t.Logf("Voucher: %v", voucher)
}
