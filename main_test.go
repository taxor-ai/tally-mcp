package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/taxor-ai/tally-mcp/pkg/logger"
	"github.com/taxor-ai/tally-mcp/pkg/mcp"
	"github.com/taxor-ai/tally-mcp/pkg/tally"
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

func TestServeMCPGETReturns405(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/mcp")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", resp.StatusCode)
	}
}

func TestServeMCPInitialize(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %q", ct)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	inner, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'result' field in response, got: %v", result)
	}
	if inner["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", inner["protocolVersion"])
	}
}

func TestServeMCPNotificationsInitialized(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected 202, got %d", resp.StatusCode)
	}
}

func TestServeMCPUnknownMethod(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"unknown/method","params":{}}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 (JSON-RPC errors use 200), got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if _, hasError := result["error"]; !hasError {
		t.Errorf("Expected 'error' field in response for unknown method, got: %v", result)
	}
}

func TestServeMCPInvalidJSON(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{invalid json`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if _, hasError := result["error"]; !hasError {
		t.Errorf("Expected 'error' field for invalid JSON, got: %v", result)
	}
}

func TestServeMCPInvalidJSONRPCVersion(t *testing.T) {
	log, _ := logger.New("warn", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, nil, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"1.0","id":1,"method":"initialize","params":{}}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if _, hasError := result["error"]; !hasError {
		t.Errorf("Expected 'error' field for invalid JSONRPC version, got: %v", result)
	}
}

func TestServeMCPToolsList(t *testing.T) {
	// Skip if tools directory is not available (CI without tools)
	toolsDir := "tools"
	if _, err := os.Stat(toolsDir); err != nil {
		t.Skipf("tools directory not found at %q, skipping tools/list integration test", toolsDir)
	}

	registry, err := tally.LoadRegistry(toolsDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}

	log, _ := logger.New("warn", "")
	client := tally.NewClient("localhost", 9000, 5)
	handler := mcp.NewHandler(client, registry, log)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMCP(w, r, handler, log)
	}))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	resp, err := http.Post(srv.URL+"/mcp", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	innerResult, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'result' field, got: %v", result)
	}
	tools, ok := innerResult["tools"].([]interface{})
	if !ok {
		t.Fatalf("Expected 'tools' array in result, got: %v", innerResult)
	}
	if len(tools) == 0 {
		t.Error("Expected at least one tool in tools/list response")
	}
}
