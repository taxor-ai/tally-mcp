package mcp

import (
	"fmt"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
)

// Handler processes MCP tool calls using the tool registry.
type Handler struct {
	client   *tally.Client
	registry *tally.Registry
	log      *logger.Logger
}

// NewHandler creates a new MCP handler backed by the given registry.
func NewHandler(client *tally.Client, registry *tally.Registry, log *logger.Logger) *Handler {
	return &Handler{client: client, registry: registry, log: log}
}

// ListTools returns all tools registered in the registry as MCP Tool structs.
func (h *Handler) ListTools() []Tool {
	defs := h.registry.All()
	tools := make([]Tool, 0, len(defs))
	for _, def := range defs {
		tools = append(tools, Tool{
			Name:        def.Name,
			Description: def.Description,
			InputSchema: def.InputSchema.ToMap(),
		})
	}
	return tools
}

// HandleToolCall dispatches a tool call generically:
//  1. Look up tool in registry
//  2. Build template params (implicit from client config + explicit from call)
//  3. Render request.xml template
//  4. POST to Tally
//  5. Apply parser.yaml spec and return result
func (h *Handler) HandleToolCall(toolName string, params map[string]interface{}) (interface{}, error) {
	def := h.registry.Get(toolName)
	if def == nil {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	if h.log != nil {
		h.log.Infof("%s called", toolName)
	}

	// Build params map with support for complex data structures
	templateParams := make(map[string]interface{}, len(params)+2)

	// Inject implicit params from client config
	for _, implicit := range def.ImplicitParams {
		switch implicit {
		case "company_name":
			templateParams["company_name"] = h.client.Company
		}
	}

	// Merge caller-supplied params (preserve original types for complex structures)
	for k, v := range params {
		templateParams[k] = v
	}

	// Render template
	rendered := tally.RenderTemplate(def.RequestXML, templateParams)

	// Log the rendered XML being sent (use Infof for visibility)
	if h.log != nil {
		h.log.Infof("XML for %s:\n%s", toolName, rendered)
	}

	// POST to Tally
	xmlResp, err := h.client.ExecuteXML(rendered)
	if err != nil {
		if h.log != nil {
			h.log.Warnf("%s failed: %v", toolName, err)
		}
		return nil, fmt.Errorf("tally request failed: %w", err)
	}

	// Log raw response
	if h.log != nil {
		h.log.Infof("Raw response from %s: %s", toolName, string(xmlResp))
	}

	// Parse response
	result, err := tally.ParseResponse(xmlResp, def.Parser)
	if err != nil {
		if h.log != nil {
			h.log.Warnf("%s parse error: %v", toolName, err)
		}
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if h.log != nil {
		h.log.Infof("%s completed", toolName)
	}
	return result, nil
}
