package tally

import (
	"fmt"
	"net/http"
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
