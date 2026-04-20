package tally

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FieldSpec describes how to extract one field from XML via XPath.
// For nested structures, set ItemsXPath to extract a list of items,
// and define Fields for each nested item.
type FieldSpec struct {
	XPath     string               `yaml:"xpath,omitempty"`
	Transform string               `yaml:"transform,omitempty"` // number, integer, boolean, tally_date
	ItemsXPath string              `yaml:"items_xpath,omitempty"` // for nested lists
	Fields    map[string]FieldSpec `yaml:"fields,omitempty"` // for nested objects/lists
}

// ParserSpec describes how to parse a Tally XML response.
// type values: "list", "object", "import_result", "raw"
type ParserSpec struct {
	Type       string               `yaml:"type"`
	ItemsXPath string               `yaml:"items_xpath,omitempty"` // for list: XPath to each item node
	RootXPath  string               `yaml:"root_xpath,omitempty"`  // for object: XPath to root node
	ResultKey  string               `yaml:"result_key,omitempty"`  // key used in response map
	Fields     map[string]FieldSpec `yaml:"fields,omitempty"`
}

// InputProperty is one parameter in a tool's input schema.
type InputProperty struct {
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Enum        []string `yaml:"enum,omitempty"`
}

// InputSchema mirrors the JSON Schema object used in MCP's tools/list response.
type InputSchema struct {
	Properties map[string]InputProperty `yaml:"properties"`
	Required   []string                 `yaml:"required,omitempty"`
}

// ToMap converts InputSchema to map[string]interface{} for MCP protocol serialisation.
func (s InputSchema) ToMap() map[string]interface{} {
	props := make(map[string]interface{})
	for name, prop := range s.Properties {
		p := map[string]interface{}{
			"type":        prop.Type,
			"description": prop.Description,
		}
		if len(prop.Enum) > 0 {
			p["enum"] = prop.Enum
		}
		props[name] = p
	}
	result := map[string]interface{}{
		"type":       "object",
		"properties": props,
	}
	if len(s.Required) > 0 {
		result["required"] = s.Required
	}
	return result
}

// ToolDefinition is a fully loaded tool: metadata + XML template + parser spec.
type ToolDefinition struct {
	Name           string      `yaml:"name"`
	Description    string      `yaml:"description"`
	ImplicitParams []string    `yaml:"implicit_params,omitempty"`
	InputSchema    InputSchema `yaml:"input_schema"`
	RequestXML     string      `yaml:"-"` // loaded from request.xml
	Parser         ParserSpec  `yaml:"-"` // loaded from parser.yaml
}

// Registry holds all registered ToolDefinitions, keyed by tool name.
type Registry struct {
	tools map[string]*ToolDefinition
	order []string // preserves registration order for tools/list
}

func newRegistry() *Registry {
	return &Registry{tools: make(map[string]*ToolDefinition)}
}

// Get returns the ToolDefinition for the given name, or nil if not found.
func (r *Registry) Get(name string) *ToolDefinition {
	return r.tools[name]
}

// All returns all ToolDefinitions in registration order.
func (r *Registry) All() []*ToolDefinition {
	out := make([]*ToolDefinition, 0, len(r.order))
	for _, name := range r.order {
		if def, ok := r.tools[name]; ok {
			out = append(out, def)
		}
	}
	return out
}

// LoadRegistry scans the templates directory on the filesystem for tool.yaml files and builds a Registry.
// Each tool lives in its own subdirectory alongside request.xml and parser.yaml.
// The templatesDir parameter should be the path to the templates directory (relative or absolute).
// For deployed binaries, this is typically {binary_dir}/templates.
// For development, it falls back to pkg/tally/templates.
func LoadRegistry(templatesDir string) (*Registry, error) {
	reg := newRegistry()

	// Resolve templates directory path
	// Priority: env var TALLY_TEMPLATES_DIR (if set) > provided path > "templates" in current dir
	var err error
	if envDir := os.Getenv("TALLY_TEMPLATES_DIR"); envDir != "" {
		templatesDir = envDir
	}

	// Check if the directory exists
	if _, err = os.Stat(templatesDir); err != nil {
		// If provided path doesn't exist, try "templates" in current directory
		if _, err2 := os.Stat("templates"); err2 == nil {
			templatesDir = "templates"
		} else {
			// Show helpful error with instructions on how to fix it
			if os.Getenv("TALLY_TEMPLATES_DIR") != "" {
				return nil, fmt.Errorf("templates directory not found at %s (from TALLY_TEMPLATES_DIR): %w", templatesDir, err)
			}
			return nil, fmt.Errorf("templates directory not found at %s or ./templates (set TALLY_TEMPLATES_DIR to override): %w", templatesDir, err)
		}
	}

	err = filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, "tool.yaml") {
			return nil
		}

		// dir is the directory containing tool.yaml (e.g., "templates/company/get_companies")
		dir := filepath.Dir(path)

		// Parse tool.yaml
		toolData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		var def ToolDefinition
		if err := yaml.Unmarshal(toolData, &def); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		// Load request.xml
		requestPath := filepath.Join(dir, "request.xml")
		requestData, err := os.ReadFile(requestPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", requestPath, err)
		}
		def.RequestXML = string(requestData)

		// Load parser.yaml (optional; default to raw)
		parserPath := filepath.Join(dir, "parser.yaml")
		if parserData, err := os.ReadFile(parserPath); err == nil {
			if err := yaml.Unmarshal(parserData, &def.Parser); err != nil {
				return fmt.Errorf("parsing %s: %w", parserPath, err)
			}
		} else {
			def.Parser = ParserSpec{Type: "raw"}
		}

		reg.tools[def.Name] = &def
		reg.order = append(reg.order, def.Name)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return reg, nil
}
