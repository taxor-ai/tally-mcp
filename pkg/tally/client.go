package tally

import (
	"bytes"
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
	Timeout int
	http    *http.Client
}

func NewClient(host string, port int, timeoutSeconds int) *Client {
	return &Client{
		Host:    host,
		Port:    port,
		Timeout: timeoutSeconds,
		http:    &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

func (c *Client) SetCompany(companyName string) {
	c.Company = companyName
}

func (c *Client) buildRPCURL() string {
	return fmt.Sprintf("http://%s:%d/", c.Host, c.Port)
}

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

// ExecuteXML posts rendered XML to Tally and returns the sanitised XML response.
func (c *Client) ExecuteXML(xmlContent string) ([]byte, error) {
	url := c.buildRPCURL()
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(xmlContent)))
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("request creation failed: %v", err))
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("failed to read response: %v", err))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
	}
	return sanitizeXML(body), nil
}

func sanitizeXML(data []byte) []byte {
	data = bytes.Map(func(r rune) rune {
		if (r >= 0x20 && r <= 0xD7FF) ||
			r == 0x09 || r == 0x0A || r == 0x0D ||
			(r >= 0xE000 && r <= 0xFFFD) ||
			(r >= 0x10000 && r <= 0x10FFFF) {
			return r
		}
		return -1
	}, data)
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
		if err == nil && val < 32 && val != 9 && val != 10 && val != 13 {
			return []byte("")
		}
		return match
	})
}
