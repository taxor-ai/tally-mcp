package tally_test

import (
	"os"
	"testing"
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

func TestRegistryLoadsBuiltinTools(t *testing.T) {
	// Set templates directory for test
	templatesDir := findTemplatesDir(t)
	os.Setenv("TALLY_TEMPLATES_DIR", templatesDir)
	defer os.Unsetenv("TALLY_TEMPLATES_DIR")

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
	os.Setenv("TALLY_TEMPLATES_DIR", templatesDir)
	defer os.Unsetenv("TALLY_TEMPLATES_DIR")

	registry, err := tally.LoadRegistry(templatesDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}
	if registry.Get("nonexistent_tool") != nil {
		t.Fatal("expected nil for unknown tool")
	}
}
