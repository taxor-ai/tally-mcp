package mcp

import (
	"fmt"

	"github.com/taxor-ai/tally-mcp/logger"
	"github.com/taxor-ai/tally-mcp/tally"
)

// Handler processes MCP tool calls
type Handler struct {
	client *tally.Client
	log    *logger.Logger
}

// NewHandler creates a new MCP handler
func NewHandler(client *tally.Client, log *logger.Logger) *Handler {
	return &Handler{
		client: client,
		log:    log,
	}
}

// HandleToolCall routes tool calls to appropriate handlers
func (h *Handler) HandleToolCall(toolName string, params map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "get_companies":
		return h.handleGetCompanies()
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// handleGetCompanies fetches all companies from Tally
func (h *Handler) handleGetCompanies() (interface{}, error) {
	if h.log != nil {
		h.log.Info("get_companies called")
	}

	// Load and execute the template
	_, err := tally.LoadTemplate("company/get_companies", map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// In a real implementation, we'd call Tally via XML-RPC here
	// For now, return a mock response for validation
	// TODO: Implement actual Tally XML-RPC call

	// Parse response and extract companies
	companies := []tally.Company{
		{Name: "DemoCompany", GUID: "comp001"},
	}

	response := map[string]interface{}{
		"success":   true,
		"companies": companies,
	}

	if h.log != nil {
		h.log.Info("get_companies completed", "count", len(companies))
	}

	return response, nil
}
