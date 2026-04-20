package tally

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// RenderTemplate fills {{ key | e }} and {{key}} placeholders in a template string.
// Also supports Go template syntax for complex data structures like lists.
func RenderTemplate(templateStr string, params map[string]interface{}) string {
	// Try Go template rendering first (for loops, conditionals, etc.)
	tmpl, err := template.New("xml").Funcs(template.FuncMap{
		"e": escapeXML, // XML escape filter like Jinja2
	}).Parse(templateStr)
	if err == nil {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err == nil {
			return buf.String()
		}
	}

	// Fallback to simple string replacement for compatibility
	result := templateStr
	for key, value := range params {
		var strValue string
		if v, ok := value.(string); ok {
			strValue = v
		} else if value != nil {
			strValue = fmt.Sprintf("%v", value)
		}
		result = strings.ReplaceAll(result, fmt.Sprintf("{{ %s | e }}", key), strValue)
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), strValue)
	}
	return result
}

// escapeXML escapes XML special characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// LoadTemplate loads and parameterizes an XML template from the templates directory
func LoadTemplate(templatePath string, params map[string]interface{}) (string, error) {
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
