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

	// Execute the template against Tally
	xmlResponse, err := h.client.ExecuteTemplate("company/get_companies", map[string]string{})
	if err != nil {
		if h.log != nil {
			h.log.Warn("get_companies failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Parse the XML response
	companies, err := tally.ParseCompaniesResponse(xmlResponse)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse companies response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
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
