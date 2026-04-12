package tally

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed templates/*/*.xml
var templatesFS embed.FS

// LoadTemplate loads and parameterizes an XML template
func LoadTemplate(templatePath string, params map[string]string) (string, error) {
	data, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s.xml", templatePath))
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}

	xml := string(data)
	for key, value := range params {
		// Handle Jinja2-style placeholders with filters: {{ name | e }}
		placeholder := fmt.Sprintf("{{ %s | e }}", key)
		xml = strings.ReplaceAll(xml, placeholder, value)

		// Also handle simple placeholders: {{name}}
		simplePlaceholder := fmt.Sprintf("{{%s}}", key)
		xml = strings.ReplaceAll(xml, simplePlaceholder, value)
	}

	return xml, nil
}
