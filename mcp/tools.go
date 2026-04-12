package mcp

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	IsWrite     bool
}

// GetCompaniesTool returns the get_companies tool definition
func GetCompaniesTool() Tool {
	return Tool{
		Name:        "get_companies",
		Description: "List all companies available in Tally",
		IsWrite:     false,
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}
}

// AllTools returns all available tools (for now, just get_companies)
func AllTools() []Tool {
	return []Tool{
		GetCompaniesTool(),
	}
}
