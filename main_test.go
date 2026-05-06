package main

import (
	"testing"
)

func TestBuildInitializeResponse(t *testing.T) {
	req := MCPRequest{JSONRPC: "2.0", ID: float64(1), Method: "initialize"}
	resp := buildInitializeResponse(req)

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got %q", resp.JSONRPC)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected Result to be map[string]interface{}, got %T", resp.Result)
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected serverInfo to be map[string]interface{}")
	}
	if serverInfo["name"] != "Tally MCP" {
		t.Errorf("Expected serverInfo.name 'Tally MCP', got %v", serverInfo["name"])
	}
}
