package tally

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("localhost", 9900, 30)
	if client == nil {
		t.Error("Expected client, got nil")
	}
	if client.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", client.Host)
	}
	if client.Port != 9900 {
		t.Errorf("Expected port 9900, got %d", client.Port)
	}
}

func TestBuildRPCURL(t *testing.T) {
	client := NewClient("localhost", 9900, 30)
	url := client.buildRPCURL()
	expected := "http://localhost:9900/"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}
