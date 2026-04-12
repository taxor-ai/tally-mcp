package tally

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// Client wraps Tally XML-RPC communication
type Client struct {
	Host    string
	Port    int
	Company string
	Timeout int // seconds
	http    *http.Client
}

// NewClient creates a new Tally client
func NewClient(host string, port int, timeoutSeconds int) *Client {
	return &Client{
		Host:    host,
		Port:    port,
		Timeout: timeoutSeconds,
		http: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// SetCompany sets the working company context
func (c *Client) SetCompany(companyName string) {
	c.Company = companyName
}

// buildRPCURL constructs the Tally RPC endpoint URL
func (c *Client) buildRPCURL() string {
	return fmt.Sprintf("http://%s:%d/", c.Host, c.Port)
}

// Ping tests connectivity to Tally
func (c *Client) Ping() error {
	url := c.buildRPCURL()
	resp, err := c.http.Get(url)
	if err != nil {
		return NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	return nil
}

// ExecuteTemplate executes a TDL template against Tally and returns the raw XML response
func (c *Client) ExecuteTemplate(templateName string, params map[string]string) ([]byte, error) {
	// Load the template
	templateContent, err := LoadTemplate(templateName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Build the RPC URL
	url := c.buildRPCURL()

	// Create the request with the template as the body
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(templateContent)))
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("request creation failed: %v", err))
	}

	// Set required headers for Tally XML-RPC (matching reference implementation)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	// Execute the request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("failed to read response: %v", err))
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
	}

	// Sanitize XML response to remove illegal XML characters
	return sanitizeXML(body), nil
}

// sanitizeXML removes illegal XML characters and entities from the byte slice
// This matches the reference implementation's approach
func sanitizeXML(data []byte) []byte {
	// Remove raw illegal characters
	data = bytes.Map(func(r rune) rune {
		if (r >= 0x20 && r <= 0xD7FF) ||
			r == 0x09 || r == 0x0A || r == 0x0D ||
			(r >= 0xE000 && r <= 0xFFFD) ||
			(r >= 0x10000 && r <= 0x10FFFF) {
			return r
		}
		return -1
	}, data)

	// Remove numeric entities for control characters
	controlEntityRegex := regexp.MustCompile(`&#(x?[0-9a-fA-F]+);`)
	return controlEntityRegex.ReplaceAllFunc(data, func(match []byte) []byte {
		entity := string(match)
		var val int
		var err error
		if len(entity) > 3 && entity[2] == 'x' {
			_, err = fmt.Sscanf(entity, "&#x%x;", &val)
		} else {
			_, err = fmt.Sscanf(entity, "&#%d;", &val)
		}

		if err == nil {
			// Check if it's a control character entity
			if val < 32 && val != 9 && val != 10 && val != 13 {
				return []byte("")
			}
		}
		return match
	})
}

// ParseCompaniesResponse parses the XML response from Tally's GetCompanies query
func ParseCompaniesResponse(xmlData []byte) ([]Company, error) {
	// Define the response structure based on Tally's XML format
	// Structure: ENVELOPE > BODY > DATA > COLLECTION > COMPANY
	type CompanyXML struct {
		Name string `xml:"NAME,attr"`
		GUID string `xml:"GUID"`
	}

	type TallyResponse struct {
		Companies []CompanyXML `xml:"BODY>DATA>COLLECTION>COMPANY"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	// Convert to Company objects
	companies := make([]Company, len(result.Companies))
	for i, c := range result.Companies {
		companies[i] = Company{
			Name: c.Name,
			GUID: c.GUID,
		}
	}

	return companies, nil
}

// ParseLedgersResponse parses the XML response from Tally's GetLedgers query
func ParseLedgersResponse(xmlData []byte, filterType string) ([]Ledger, error) {
	type LedgerXML struct {
		Name     string `xml:"NAME,attr"`
		Parent   string `xml:"PARENT"`
		IsAggregate string `xml:"$ISAGGREGATE"`
		ReservedName string `xml:"$RESERVEDNAME"`
	}

	type TallyResponse struct {
		Ledgers []LedgerXML `xml:"BODY>DATA>COLLECTION>LEDGER"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	ledgers := make([]Ledger, 0)
	for _, l := range result.Ledgers {
		ledger := Ledger{
			Name:        l.Name,
			ParentGroup: l.Parent,
		}
		ledgers = append(ledgers, ledger)
	}

	return ledgers, nil
}

// ParseLedgerDetailsResponse parses the XML response for a specific ledger
func ParseLedgerDetailsResponse(xmlData []byte) (*Ledger, error) {
	type LedgerXML struct {
		Name        string `xml:"NAME"`
		Parent      string `xml:"PARENT"`
		Description string `xml:"DESCRIPTION"`
		CreatedDate string `xml:"CREATEDDATE"`
		Balance     string `xml:"BALANCE"`
	}

	type TallyResponse struct {
		Ledgers []LedgerXML `xml:"BODY>DATA>COLLECTION>LEDGER"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	if len(result.Ledgers) == 0 {
		return nil, fmt.Errorf("ledger not found")
	}

	l := result.Ledgers[0]
	balance := 0.0
	if b, err := strconv.ParseFloat(l.Balance, 64); err == nil {
		balance = b
	}

	return &Ledger{
		Name:        l.Name,
		ParentGroup: l.Parent,
		Description: l.Description,
		CreatedDate: l.CreatedDate,
		Balance:     balance,
	}, nil
}

// ParseDebtorsResponse parses the XML response from Tally's GetDebtors query
func ParseDebtorsResponse(xmlData []byte) ([]Debtor, error) {
	type DebtorXML struct {
		Name              string `xml:"NAME,attr"`
		OutstandingAmount string `xml:"OUTSTANDINGAMOUNT"`
		CreditLimit       string `xml:"CREDITLIMIT"`
		DaysOutstanding   string `xml:"DAYSOUTSTANDING"`
	}

	type TallyResponse struct {
		Debtors []DebtorXML `xml:"BODY>DATA>COLLECTION>LEDGER"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	debtors := make([]Debtor, 0)
	for _, d := range result.Debtors {
		outstanding := 0.0
		if o, err := strconv.ParseFloat(d.OutstandingAmount, 64); err == nil {
			outstanding = o
		}
		creditLimit := 0.0
		if c, err := strconv.ParseFloat(d.CreditLimit, 64); err == nil {
			creditLimit = c
		}
		daysOut := 0
		if do, err := strconv.Atoi(d.DaysOutstanding); err == nil {
			daysOut = do
		}

		debtors = append(debtors, Debtor{
			Name:              d.Name,
			OutstandingAmount: outstanding,
			CreditLimit:       creditLimit,
			DaysOutstanding:   daysOut,
		})
	}

	return debtors, nil
}

// ParseCreditorsResponse parses the XML response from Tally's GetCreditors query
func ParseCreditorsResponse(xmlData []byte) ([]Creditor, error) {
	type CreditorXML struct {
		Name              string `xml:"NAME,attr"`
		OutstandingAmount string `xml:"OUTSTANDINGAMOUNT"`
		PaymentTerms      string `xml:"PAYMENTTERMS"`
		DaysOutstanding   string `xml:"DAYSOUTSTANDING"`
	}

	type TallyResponse struct {
		Creditors []CreditorXML `xml:"BODY>DATA>COLLECTION>LEDGER"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	creditors := make([]Creditor, 0)
	for _, c := range result.Creditors {
		outstanding := 0.0
		if o, err := strconv.ParseFloat(c.OutstandingAmount, 64); err == nil {
			outstanding = o
		}
		daysOut := 0
		if do, err := strconv.Atoi(c.DaysOutstanding); err == nil {
			daysOut = do
		}

		creditors = append(creditors, Creditor{
			Name:              c.Name,
			OutstandingAmount: outstanding,
			PaymentTerms:      c.PaymentTerms,
			DaysOutstanding:   daysOut,
		})
	}

	return creditors, nil
}

// ParseVouchersResponse parses the XML response from Tally's GetVouchers query
func ParseVouchersResponse(xmlData []byte) ([]Voucher, error) {
	type VoucherXML struct {
		Date            string `xml:"DATE"`
		VoucherNumber   string `xml:"VOUCHERNUMBER"`
		Reference       string `xml:"REFERENCE"`
		Narration       string `xml:"NARRATION"`
		VoucherTypeName string `xml:"VOUCHERTYPENAME"`
		Amount          string `xml:"AMOUNT"`
		Status          string `xml:"STATUS"`
	}

	type TallyResponse struct {
		Vouchers []VoucherXML `xml:"BODY>DATA>COLLECTION>VOUCHER"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	vouchers := make([]Voucher, 0)
	for _, v := range result.Vouchers {
		amount := 0.0
		if a, err := strconv.ParseFloat(v.Amount, 64); err == nil {
			amount = a
		}

		vouchers = append(vouchers, Voucher{
			VoucherID:   v.VoucherNumber,
			Type:        v.VoucherTypeName,
			Date:        v.Date,
			Amount:      amount,
			Status:      v.Status,
			Description: v.Narration,
		})
	}

	return vouchers, nil
}

// ParseCreateResponse parses the response from create operations (ledger/voucher)
func ParseCreateResponse(xmlData []byte) (bool, error) {
	type TallyResponse struct {
		Status string `xml:"STATUS"`
		Error  string `xml:"ERROR"`
	}

	var result TallyResponse
	err := xml.Unmarshal(xmlData, &result)
	if err != nil {
		return false, fmt.Errorf("failed to parse XML response: %w", err)
	}

	// Check if there was an error
	if result.Error != "" {
		return false, fmt.Errorf("tally error: %s", result.Error)
	}

	return result.Status == "Success", nil
}
