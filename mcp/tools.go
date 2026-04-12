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
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}

// GetLedgersTool returns the get_ledgers tool definition
func GetLedgersTool() Tool {
	return Tool{
		Name:        "get_ledgers",
		Description: "List all ledgers in Tally, optionally filtered by type",
		IsWrite:     false,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter_type": map[string]interface{}{
					"type":        "string",
					"description": "Filter ledgers by type: asset, liability, income, expense, debtor, creditor, or all",
					"enum":        []string{"asset", "liability", "income", "expense", "debtor", "creditor", "all"},
				},
			},
		},
	}
}

// GetLedgerDetailsTool returns the get_ledger_details tool definition
func GetLedgerDetailsTool() Tool {
	return Tool{
		Name:        "get_ledger_details",
		Description: "Get detailed information for a specific ledger",
		IsWrite:     false,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ledger_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the ledger to query",
				},
			},
			"required": []string{"ledger_name"},
		},
	}
}

// GetDebtorsTool returns the get_debtors tool definition
func GetDebtorsTool() Tool {
	return Tool{
		Name:        "get_debtors",
		Description: "List all debtor ledgers with outstanding amounts",
		IsWrite:     false,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}

// GetCreditorsTool returns the get_creditors tool definition
func GetCreditorsTool() Tool {
	return Tool{
		Name:        "get_creditors",
		Description: "List all creditor ledgers with outstanding amounts",
		IsWrite:     false,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}

// CreateLedgerTool returns the create_ledger tool definition
func CreateLedgerTool() Tool {
	return Tool{
		Name:        "create_ledger",
		Description: "Create a new ledger account in Tally",
		IsWrite:     true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Ledger name",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Ledger type: Asset, Liability, Income, Expense, Debtor, Creditor",
					"enum":        []string{"Asset", "Liability", "Income", "Expense", "Debtor", "Creditor"},
				},
				"parent_group": map[string]interface{}{
					"type":        "string",
					"description": "Parent group name (optional)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Ledger description (optional)",
				},
				"opening_balance": map[string]interface{}{
					"type":        "number",
					"description": "Opening balance amount (optional, default: 0)",
				},
			},
			"required": []string{"name", "type"},
		},
	}
}

// CreateVoucherTool returns the create_voucher tool definition
func CreateVoucherTool() Tool {
	return Tool{
		Name:        "create_voucher",
		Description: "Create a new voucher (invoice, payment, or expense entry) in Tally",
		IsWrite:     true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"voucher_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of voucher: Invoice, Payment, or Expense",
					"enum":        []string{"Invoice", "Payment", "Expense"},
				},
				"date": map[string]interface{}{
					"type":        "string",
					"description": "Voucher date in YYYY-MM-DD format",
				},
				"reference_number": map[string]interface{}{
					"type":        "string",
					"description": "Reference number (e.g., invoice number, check number)",
				},
				"party_name": map[string]interface{}{
					"type":        "string",
					"description": "Party name (for invoices/payments, must be a debtor/creditor)",
				},
				"line_items": map[string]interface{}{
					"type":        "array",
					"description": "Array of line items",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"ledger_name": map[string]interface{}{
								"type":        "string",
								"description": "Ledger account name",
							},
							"amount": map[string]interface{}{
								"type":        "number",
								"description": "Amount for this line item",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Line item description (optional)",
							},
						},
						"required": []string{"ledger_name", "amount"},
					},
				},
				"notes": map[string]interface{}{
					"type":        "string",
					"description": "Additional notes (optional)",
				},
			},
			"required": []string{"voucher_type", "date", "line_items"},
		},
	}
}

// AllTools returns all available tools
func AllTools() []Tool {
	return []Tool{
		GetCompaniesTool(),
		GetLedgersTool(),
		GetLedgerDetailsTool(),
		GetDebtorsTool(),
		GetCreditorsTool(),
		CreateLedgerTool(),
		CreateVoucherTool(),
	}
}
