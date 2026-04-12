package tally

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
		Name string `xml:"NAME"`
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
