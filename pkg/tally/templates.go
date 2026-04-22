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
// Strategy: first replace scalar top-level params, then run Go template for loops/conditionals.
func RenderTemplate(templateStr string, params map[string]interface{}) string {
	// Step 1: Replace scalar top-level params using simple string replacement.
	// This handles {{ key | e }} and {{key}} patterns used in get/create tools.
	result := templateStr
	for key, value := range params {
		var strValue string
		if v, ok := value.(string); ok {
			strValue = escapeXML(v)
		} else if value != nil {
			switch value.(type) {
			case []interface{}, []map[string]interface{}:
				continue // skip complex types — handled by Go template below
			default:
				strValue = fmt.Sprintf("%v", value)
			}
		}
		result = strings.ReplaceAll(result, fmt.Sprintf("{{ %s | e }}", key), strValue)
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), strValue)
	}

	// Step 2: Run Go template on the result to handle range/if over complex types.
	tmpl, err := template.New("xml").Funcs(template.FuncMap{
		"e": escapeXML,
		"isNeg": func(v interface{}) bool {
			switch n := v.(type) {
			case int:     return n < 0
			case int64:   return n < 0
			case float32: return n < 0
			case float64: return n < 0
			}
			return false
		},
	}).Parse(result)
	if err == nil {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err == nil {
			return buf.String()
		}
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

// LoadTemplate loads and parameterizes an XML template from the tools directory
func LoadTemplate(templatePath string, params map[string]interface{}) (string, error) {
	// Load from tools/{category}/{toolname}/request.xml
	// Example: "ledger/create_ledger" -> "tools/ledger/create_ledger/request.xml"
	fullPath := fmt.Sprintf("tools/%s/request.xml", templatePath)
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
