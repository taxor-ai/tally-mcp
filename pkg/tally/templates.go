package tally

import (
	"fmt"
	"os"
	"strings"
)

// RenderTemplate fills {{ key | e }} and {{key}} placeholders in a template string.
func RenderTemplate(templateStr string, params map[string]string) string {
	result := templateStr
	for key, value := range params {
		result = strings.ReplaceAll(result, fmt.Sprintf("{{ %s | e }}", key), value)
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), value)
	}
	return result
}

// LoadTemplate loads and parameterizes an XML template from the templates directory
func LoadTemplate(templatePath string, params map[string]string) (string, error) {
	// Load from templates/{category}/{toolname}/request.xml
	// Example: "ledger/create_ledger" -> "templates/ledger/create_ledger/request.xml"
	fullPath := fmt.Sprintf("templates/%s/request.xml", templatePath)
	data, err := readFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}
	return RenderTemplate(string(data), params), nil
}

// readFile is a simple wrapper for os.ReadFile (can be swapped for embed.FS if needed)
func readFile(path string) ([]byte, error) {
	return readFileOS(path)
}

// readFileOS reads from the filesystem
func readFileOS(path string) ([]byte, error) {
	return os.ReadFile(path)
}
