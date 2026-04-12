package tally_test

import (
	"testing"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

var sampleCompaniesXML = []byte(`<ENVELOPE>
  <BODY><DATA><COLLECTION>
    <COMPANY NAME="Alpha Corp"><GUID>guid-001</GUID></COMPANY>
    <COMPANY NAME="Beta Ltd"><GUID>guid-002</GUID></COMPANY>
  </COLLECTION></DATA></BODY>
</ENVELOPE>`)

func TestParseList(t *testing.T) {
	spec := tally.ParserSpec{
		Type:       "list",
		ItemsXPath: "//COLLECTION/COMPANY",
		ResultKey:  "companies",
		Fields: map[string]tally.FieldSpec{
			"name": {XPath: "@NAME"},
			"guid": {XPath: "GUID"},
		},
	}
	result, err := tally.ParseResponse(sampleCompaniesXML, spec)
	if err != nil {
		t.Fatalf("ParseResponse error: %v", err)
	}
	items, ok := result["companies"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected []map[string]interface{}, got %T", result["companies"])
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0]["name"] != "Alpha Corp" {
		t.Errorf("expected name=Alpha Corp, got %v", items[0]["name"])
	}
	if items[0]["guid"] != "guid-001" {
		t.Errorf("expected guid=guid-001, got %v", items[0]["guid"])
	}
}

var sampleLedgerXML = []byte(`<ENVELOPE>
  <BODY><DATA><COLLECTION>
    <LEDGER NAME="Cash">
      <PARENT>Current Assets</PARENT>
      <CLOSINGBALANCE>5000.00</CLOSINGBALANCE>
    </LEDGER>
  </COLLECTION></DATA></BODY>
</ENVELOPE>`)

func TestParseObject(t *testing.T) {
	spec := tally.ParserSpec{
		Type:      "object",
		RootXPath: "//COLLECTION/LEDGER[1]",
		ResultKey: "ledger",
		Fields: map[string]tally.FieldSpec{
			"name":    {XPath: "@NAME"},
			"parent":  {XPath: "PARENT"},
			"balance": {XPath: "CLOSINGBALANCE", Transform: "number"},
		},
	}
	result, err := tally.ParseResponse(sampleLedgerXML, spec)
	if err != nil {
		t.Fatalf("ParseResponse error: %v", err)
	}
	ledger, ok := result["ledger"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result["ledger"])
	}
	if ledger["name"] != "Cash" {
		t.Errorf("expected name=Cash, got %v", ledger["name"])
	}
	if ledger["balance"] != 5000.0 {
		t.Errorf("expected balance=5000.0, got %v", ledger["balance"])
	}
}

var sampleImportResultXML = []byte(`<ENVELOPE>
  <BODY><DATA><IMPORTRESULT>
    <CREATED>1</CREATED><ALTERED>0</ALTERED><DELETED>0</DELETED>
  </IMPORTRESULT></DATA></BODY>
</ENVELOPE>`)

func TestParseImportResult(t *testing.T) {
	spec := tally.ParserSpec{Type: "import_result"}
	result, err := tally.ParseResponse(sampleImportResultXML, spec)
	if err != nil {
		t.Fatalf("ParseResponse error: %v", err)
	}
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["created"] != 1 {
		t.Errorf("expected created=1, got %v", result["created"])
	}
}

func TestTransformTallyDate(t *testing.T) {
	spec := tally.ParserSpec{
		Type:       "list",
		ItemsXPath: "//VOUCHER",
		Fields: map[string]tally.FieldSpec{
			"date": {XPath: "DATE", Transform: "tally_date"},
		},
	}
	xml := []byte(`<ENVELOPE><BODY><DATA><COLLECTION>
        <VOUCHER><DATE>20240401</DATE></VOUCHER>
    </COLLECTION></DATA></BODY></ENVELOPE>`)
	result, err := tally.ParseResponse(xml, spec)
	if err != nil {
		t.Fatalf("ParseResponse error: %v", err)
	}
	items := result["items"].([]map[string]interface{})
	if items[0]["date"] != "2024-04-01" {
		t.Errorf("expected 2024-04-01, got %v", items[0]["date"])
	}
}
