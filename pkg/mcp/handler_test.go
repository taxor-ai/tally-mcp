package mcp

import (
	"os"
	"testing"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

// findTemplatesDir finds the templates directory from expected paths
func findTemplatesDir(t *testing.T) string {
	candidates := []string{
		"pkg/tally/templates",
		"../../pkg/tally/templates",
		"../../../pkg/tally/templates",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Fatal("Could not find pkg/tally/templates from any expected path")
	return ""
}

func TestHandleGetCompanies(t *testing.T) {
	// Skip if Tally is not available
	if os.Getenv("TALLY_HOST") == "" {
		t.Skip("TALLY_HOST not set, skipping test that requires Tally")
	}

	host := os.Getenv("TALLY_HOST")
	company := os.Getenv("TALLY_COMPANY")
	client := tally.NewClient(host, 9900, 30)
	client.SetCompany(company)

	log, _ := logger.New("warn", "")
	templatesDir := findTemplatesDir(t)
	os.Setenv("TALLY_TEMPLATES_DIR", templatesDir)
	defer os.Unsetenv("TALLY_TEMPLATES_DIR")
	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	handler := NewHandler(client, registry, log)

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
	log, _ := logger.New("warn", "")
	templatesDir := findTemplatesDir(t)
	os.Setenv("TALLY_TEMPLATES_DIR", templatesDir)
	defer os.Unsetenv("TALLY_TEMPLATES_DIR")
	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	handler := NewHandler(client, registry, log)

	_, err2 := handler.HandleToolCall("unknown_tool", map[string]interface{}{})

	if err2 == nil {
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
