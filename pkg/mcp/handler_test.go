package mcp

import (
	"os"
	"testing"

	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

func TestHandleGetCompanies(t *testing.T) {
	// Skip if Tally is not available
	if os.Getenv("TALLY_HOST") == "" {
		t.Skip("TALLY_HOST not set, skipping test that requires Tally")
	}

	host := os.Getenv("TALLY_HOST")
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, 9900, 30)
	client.SetCompany(company)
	handler := NewHandler(client, nil)

	result, err := handler.HandleToolCall("get_companies", map[string]interface{}{})

	if err != nil {
		t.Fatalf("HandleToolCall failed: %v", err)
	}

	if result == nil {
		t.Error("Expected result, got nil")
	}

	// Check response structure
	response, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("Expected map[string]interface{}, got %T", result)
	}

	success, ok := response["success"].(bool)
	if !ok || !success {
		t.Error("Expected success=true in response")
	}
}

func TestUnknownTool(t *testing.T) {
	client := tally.NewClient("localhost", 9900, 30)
	handler := NewHandler(client, nil)

	_, err := handler.HandleToolCall("unknown_tool", map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for unknown tool")
	}
}

func TestMCPResponse(t *testing.T) {
	resp := NewSuccessResponse(map[string]string{"test": "data"})

	if !resp.Success {
		t.Error("Expected Success=true")
	}

	jsonStr, err := resp.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON")
	}
}
