package tally

// Ledger represents a chart of accounts entry
type Ledger struct {
	Name        string
	Type        string // Asset, Liability, Income, Expense, Debtor, Creditor
	ParentGroup string
	Balance     float64
	Description string
	CreatedDate string
}

// Debtor represents a debtor ledger with outstanding amount
type Debtor struct {
	Name                 string
	OutstandingAmount    float64
	CreditLimit          float64
	DaysOutstanding      int
}

// Creditor represents a creditor ledger with outstanding amount
type Creditor struct {
	Name                string
	OutstandingAmount   float64
	PaymentTerms        string
	DaysOutstanding     int
}

// Voucher represents an accounting voucher (invoice, payment, expense)
type Voucher struct {
	VoucherID   string
	Type        string // Invoice, Payment, Expense
	Date        string
	Party       string
	Amount      float64
	Status      string
	Description string
}

// LineItem represents a line in a voucher
type LineItem struct {
	LedgerName  string
	Amount      float64
	Description string
}

// CreateLedgerRequest represents a request to create a ledger
type CreateLedgerRequest struct {
	Name           string
	Type           string
	ParentGroup    string
	Description    string
	OpeningBalance float64
}

// CreateVoucherRequest represents a request to create a voucher
type CreateVoucherRequest struct {
	VoucherType    string
	Date           string
	ReferenceNum   string
	PartyName      string
	LineItems      []LineItem
	Notes          string
}

// TallyResponse is the base response from Tally
type TallyResponse struct {
	Success bool
	Message string
	Error   string
	Data    interface{}
}

// Company represents a Tally company
type Company struct {
	Name string
	GUID string
}
