package mcp

import (
	"fmt"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
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
	case "get_ledgers":
		return h.handleGetLedgers(params)
	case "get_ledger_details":
		return h.handleGetLedgerDetails(params)
	case "get_debtors":
		return h.handleGetDebtors()
	case "get_creditors":
		return h.handleGetCreditors()
	case "create_ledger":
		return h.handleCreateLedger(params)
	case "create_voucher":
		return h.handleCreateVoucher(params)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// handleGetCompanies fetches all companies from Tally
func (h *Handler) handleGetCompanies() (interface{}, error) {
	if h.log != nil {
		h.log.Info("get_companies called")
	}

	xmlResponse, err := h.client.ExecuteTemplate("company/get_companies", map[string]string{})
	if err != nil {
		if h.log != nil {
			h.log.Warn("get_companies failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

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

// handleGetLedgers fetches all ledgers, optionally filtered by type
func (h *Handler) handleGetLedgers(params map[string]interface{}) (interface{}, error) {
	filterType := "all"
	if f, ok := params["filter_type"].(string); ok {
		filterType = f
	}

	if h.log != nil {
		h.log.Info("get_ledgers called", "filter_type", filterType)
	}

	xmlResponse, err := h.client.ExecuteTemplate("ledger/get_ledgers", map[string]string{})
	if err != nil {
		if h.log != nil {
			h.log.Warn("get_ledgers failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	if h.log != nil {
		h.log.Debugf("get_ledgers XML response length: %d bytes", len(xmlResponse))
	}

	ledgers, err := tally.ParseLedgersResponse(xmlResponse, filterType)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse ledgers response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := map[string]interface{}{
		"success":  true,
		"ledgers":  ledgers,
		"count":    len(ledgers),
	}

	if h.log != nil {
		h.log.Info("get_ledgers completed", "count", len(ledgers), "filter_type", filterType)
	}

	return response, nil
}

// handleGetLedgerDetails fetches details for a specific ledger
func (h *Handler) handleGetLedgerDetails(params map[string]interface{}) (interface{}, error) {
	ledgerName, ok := params["ledger_name"].(string)
	if !ok {
		return nil, fmt.Errorf("ledger_name parameter is required")
	}

	if h.log != nil {
		h.log.Info("get_ledger_details called", "ledger_name", ledgerName)
	}

	xmlResponse, err := h.client.ExecuteTemplate("ledger/get_ledger_details", map[string]string{
		"ledger_name": ledgerName,
	})
	if err != nil {
		if h.log != nil {
			h.log.Warn("get_ledger_details failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	ledger, err := tally.ParseLedgerDetailsResponse(xmlResponse)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse ledger details response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if h.log != nil {
		h.log.Info("get_ledger_details completed", "ledger_name", ledgerName)
	}

	return map[string]interface{}{
		"success": true,
		"ledger":  ledger,
	}, nil
}

// handleGetDebtors fetches all debtor ledgers
func (h *Handler) handleGetDebtors() (interface{}, error) {
	if h.log != nil {
		h.log.Info("get_debtors called")
	}

	xmlResponse, err := h.client.ExecuteTemplate("debtor_creditor/get_debtors", map[string]string{})
	if err != nil {
		if h.log != nil {
			h.log.Warn("get_debtors failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	debtors, err := tally.ParseDebtorsResponse(xmlResponse)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse debtors response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := map[string]interface{}{
		"success":  true,
		"debtors":  debtors,
		"count":    len(debtors),
	}

	if h.log != nil {
		h.log.Info("get_debtors completed", "count", len(debtors))
	}

	return response, nil
}

// handleGetCreditors fetches all creditor ledgers
func (h *Handler) handleGetCreditors() (interface{}, error) {
	if h.log != nil {
		h.log.Info("get_creditors called")
	}

	xmlResponse, err := h.client.ExecuteTemplate("debtor_creditor/get_creditors", map[string]string{})
	if err != nil {
		if h.log != nil {
			h.log.Warn("get_creditors failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	if h.log != nil {
		h.log.Debugf("get_creditors XML response length: %d bytes", len(xmlResponse))
	}

	creditors, err := tally.ParseCreditorsResponse(xmlResponse)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse creditors response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := map[string]interface{}{
		"success":   true,
		"creditors": creditors,
		"count":     len(creditors),
	}

	if h.log != nil {
		h.log.Info("get_creditors completed", "count", len(creditors))
	}

	return response, nil
}

// handleCreateLedger creates a new ledger
func (h *Handler) handleCreateLedger(params map[string]interface{}) (interface{}, error) {
	req := tally.CreateLedgerRequest{}

	if name, ok := params["name"].(string); ok {
		req.Name = name
	} else {
		return nil, fmt.Errorf("name parameter is required")
	}

	if ledgerType, ok := params["type"].(string); ok {
		req.Type = ledgerType
	} else {
		return nil, fmt.Errorf("type parameter is required")
	}

	if parentGroup, ok := params["parent_group"].(string); ok {
		req.ParentGroup = parentGroup
	}

	if description, ok := params["description"].(string); ok {
		req.Description = description
	}

	if balance, ok := params["opening_balance"].(float64); ok {
		req.OpeningBalance = balance
	}

	if h.log != nil {
		h.log.Info("create_ledger called", "name", req.Name, "type", req.Type)
	}

	xmlResponse, err := h.client.ExecuteTemplate("ledger/create_ledger", map[string]string{
		"ledger_name":      req.Name,
		"ledger_type":      req.Type,
		"parent_group":     req.ParentGroup,
		"description":      req.Description,
		"opening_balance":  fmt.Sprintf("%.2f", req.OpeningBalance),
	})
	if err != nil {
		if h.log != nil {
			h.log.Warn("create_ledger failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	success, err := tally.ParseCreateResponse(xmlResponse)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse create ledger response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := map[string]interface{}{
		"success":       success,
		"ledger_name":   req.Name,
		"message":       "Ledger created successfully",
	}

	if h.log != nil {
		h.log.Info("create_ledger completed", "name", req.Name)
	}

	return response, nil
}

// handleCreateVoucher creates a new voucher
func (h *Handler) handleCreateVoucher(params map[string]interface{}) (interface{}, error) {
	req := tally.CreateVoucherRequest{}

	if vType, ok := params["voucher_type"].(string); ok {
		req.VoucherType = vType
	} else {
		return nil, fmt.Errorf("voucher_type parameter is required")
	}

	if date, ok := params["date"].(string); ok {
		req.Date = date
	} else {
		return nil, fmt.Errorf("date parameter is required")
	}

	if refNum, ok := params["reference_number"].(string); ok {
		req.ReferenceNum = refNum
	}

	if partyName, ok := params["party_name"].(string); ok {
		req.PartyName = partyName
	}

	if notes, ok := params["notes"].(string); ok {
		req.Notes = notes
	}

	// Parse line items
	if lineItemsIface, ok := params["line_items"].([]interface{}); ok {
		for _, item := range lineItemsIface {
			if itemMap, ok := item.(map[string]interface{}); ok {
				li := tally.LineItem{}
				if ledger, ok := itemMap["ledger_name"].(string); ok {
					li.LedgerName = ledger
				}
				if amount, ok := itemMap["amount"].(float64); ok {
					li.Amount = amount
				}
				if desc, ok := itemMap["description"].(string); ok {
					li.Description = desc
				}
				req.LineItems = append(req.LineItems, li)
			}
		}
	}

	if len(req.LineItems) == 0 {
		return nil, fmt.Errorf("line_items parameter is required and must not be empty")
	}

	if h.log != nil {
		h.log.Info("create_voucher called", "type", req.VoucherType, "date", req.Date, "items", len(req.LineItems))
	}

	xmlResponse, err := h.client.ExecuteTemplate("voucher/create_voucher", map[string]string{
		"voucher_type":      req.VoucherType,
		"date":              req.Date,
		"reference_number":  req.ReferenceNum,
		"party_name":        req.PartyName,
		"notes":             req.Notes,
	})
	if err != nil {
		if h.log != nil {
			h.log.Warn("create_voucher failed", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	success, err := tally.ParseCreateResponse(xmlResponse)
	if err != nil {
		if h.log != nil {
			h.log.Warn("failed to parse create voucher response", "error", err.Error())
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response := map[string]interface{}{
		"success":     success,
		"voucher_id":  req.ReferenceNum,
		"date":        req.Date,
		"message":     "Voucher created successfully",
	}

	if h.log != nil {
		h.log.Info("create_voucher completed", "type", req.VoucherType)
	}

	return response, nil
}
