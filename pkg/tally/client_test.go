package tally

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("localhost", 9900, 30)
	if client == nil {
		t.Error("Expected client, got nil")
	}
	if client.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", client.Host)
	}
	if client.Port != 9900 {
		t.Errorf("Expected port 9900, got %d", client.Port)
	}
}

func TestBuildRPCURL(t *testing.T) {
	client := NewClient("localhost", 9900, 30)
	url := client.buildRPCURL()
	expected := "http://localhost:9900/"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestSetCompany(t *testing.T) {
	client := NewClient("localhost", 9900, 30)
	client.SetCompany("TestCompany")
	if client.Company != "TestCompany" {
		t.Errorf("Expected company 'TestCompany', got '%s'", client.Company)
	}
}

func TestParseCompaniesResponse(t *testing.T) {
	// Sample XML response from Tally (matching actual structure)
	xmlResponse := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
	<BODY>
		<DATA>
			<COLLECTION>
				<COMPANY NAME="ABC Corporation">
					<GUID>guid-abc-001</GUID>
				</COMPANY>
				<COMPANY NAME="XYZ Traders">
					<GUID>guid-xyz-002</GUID>
				</COMPANY>
			</COLLECTION>
		</DATA>
	</BODY>
</ENVELOPE>`)

	spec := ParserSpec{
		Type:       "list",
		ItemsXPath: "//COLLECTION/COMPANY",
		ResultKey:  "companies",
		Fields: map[string]FieldSpec{
			"name": {XPath: "@NAME"},
			"guid": {XPath: "GUID"},
		},
	}

	result, err := ParseResponse(xmlResponse, spec)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	companies, ok := result["companies"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected []map[string]interface{}, got %T", result["companies"])
	}

	if len(companies) != 2 {
		t.Errorf("Expected 2 companies, got %d", len(companies))
	}

	if companies[0]["name"] != "ABC Corporation" {
		t.Errorf("Expected first company name 'ABC Corporation', got '%s'", companies[0]["name"])
	}

	if companies[0]["guid"] != "guid-abc-001" {
		t.Errorf("Expected first company GUID 'guid-abc-001', got '%s'", companies[0]["guid"])
	}

	if companies[1]["name"] != "XYZ Traders" {
		t.Errorf("Expected second company name 'XYZ Traders', got '%s'", companies[1]["name"])
	}
}

func TestParseCompaniesResponseEmptyList(t *testing.T) {
	// XML response with no companies (matching actual Tally structure)
	xmlResponse := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ENVELOPE>
	<BODY>
		<DATA>
			<COLLECTION>
			</COLLECTION>
		</DATA>
	</BODY>
</ENVELOPE>`)

	spec := ParserSpec{
		Type:       "list",
		ItemsXPath: "//COLLECTION/COMPANY",
		ResultKey:  "companies",
		Fields: map[string]FieldSpec{
			"name": {XPath: "@NAME"},
			"guid": {XPath: "GUID"},
		},
	}

	result, err := ParseResponse(xmlResponse, spec)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	companies, ok := result["companies"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected []map[string]interface{}, got %T", result["companies"])
	}

	if len(companies) != 0 {
		t.Errorf("Expected 0 companies, got %d", len(companies))
	}
}

func TestParseCompaniesResponseInvalidXML(t *testing.T) {
	invalidXML := []byte(`not valid xml`)

	spec := ParserSpec{
		Type:       "list",
		ItemsXPath: "//COLLECTION/COMPANY",
		ResultKey:  "companies",
		Fields: map[string]FieldSpec{
			"name": {XPath: "@NAME"},
			"guid": {XPath: "GUID"},
		},
	}

	_, err := ParseResponse(invalidXML, spec)
	if err == nil {
		t.Fatal("Expected error for invalid XML")
	}
}
