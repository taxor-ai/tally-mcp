package tally_test

import (
	"os"
	"testing"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

// findTemplatesDir finds the templates directory from expected paths
func findTemplatesDir(t *testing.T) string {
	candidates := []string{
		"tools",
		"../../tools",
		"../../../tools",
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Fatal("Could not find tools from any expected path")
	return ""
}

func TestRegistryLoadsBuiltinTools(t *testing.T) {
	// Set templates directory for test
	templatesDir := findTemplatesDir(t)
	os.Setenv("TALLY_TOOLS_DIR", templatesDir)
	defer os.Unsetenv("TALLY_TOOLS_DIR")

	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	tools := registry.All()
	if len(tools) == 0 {
		t.Fatal("expected built-in tools to be loaded")
	}
}

func TestRegistryGetUnknownTool(t *testing.T) {
	// Set templates directory for test
	templatesDir := findTemplatesDir(t)
	os.Setenv("TALLY_TOOLS_DIR", templatesDir)
	defer os.Unsetenv("TALLY_TOOLS_DIR")

	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	if registry.Get("nonexistent_tool") != nil {
		t.Fatal("expected nil for unknown tool")
	}
}
