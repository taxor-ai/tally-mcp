package tally

import (
	"testing"
)

func TestTallyError(t *testing.T) {
	err := NewTallyError(403, "Ledger already exists")
	if err.Code != 403 {
		t.Errorf("Expected code 403, got %d", err.Code)
	}
	if err.Message != "Ledger already exists" {
		t.Errorf("Expected message 'Ledger already exists', got '%s'", err.Message)
	}
}

func TestConnectionError(t *testing.T) {
	err := NewConnectionError("localhost:9900", "connection refused")
	if err.Type != "ConnectionError" {
		t.Errorf("Expected type 'ConnectionError', got '%s'", err.Type)
	}
}
