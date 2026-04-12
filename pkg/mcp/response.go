package mcp

import (
	"encoding/json"
)

// MCPResponse represents a standard MCP response
type MCPResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ToJSON converts response to JSON string
func (r MCPResponse) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) MCPResponse {
	return MCPResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) MCPResponse {
	return MCPResponse{
		Success: false,
		Error:   err.Error(),
	}
}
