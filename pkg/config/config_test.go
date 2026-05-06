package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("TALLY_HOST", "localhost")
	os.Setenv("TALLY_PORT", "9900")
	os.Setenv("TALLY_COMPANY", "DemoCompany")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Tally.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Tally.Host)
	}
	if cfg.Tally.Port != 9900 {
		t.Errorf("Expected port 9900, got %d", cfg.Tally.Port)
	}
	if cfg.Tally.Company != "DemoCompany" {
		t.Errorf("Expected company 'DemoCompany', got '%s'", cfg.Tally.Company)
	}
}

func TestValidation(t *testing.T) {
	os.Setenv("TALLY_HOST", "")
	os.Setenv("TALLY_PORT", "")
	os.Setenv("TALLY_COMPANY", "")

	_, err := Load()
	if err == nil {
		t.Error("Expected validation error for missing required fields")
	}
}

func TestHTTPConfigDefaults(t *testing.T) {
	os.Setenv("TALLY_HOST", "localhost")
	os.Setenv("TALLY_PORT", "9900")
	os.Setenv("TALLY_COMPANY", "DemoCompany")
	defer os.Unsetenv("MCP_HTTP_PORT")
	defer os.Unsetenv("MCP_HTTP_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.HTTP.Port != "" {
		t.Errorf("Expected empty HTTP port (stdio mode), got %q", cfg.HTTP.Port)
	}
	if cfg.HTTP.Host != "0.0.0.0" {
		t.Errorf("Expected default HTTP host '0.0.0.0', got %q", cfg.HTTP.Host)
	}
}

func TestHTTPConfigFromEnv(t *testing.T) {
	os.Setenv("TALLY_HOST", "localhost")
	os.Setenv("TALLY_PORT", "9900")
	os.Setenv("TALLY_COMPANY", "DemoCompany")
	os.Setenv("MCP_HTTP_PORT", "9090")
	os.Setenv("MCP_HTTP_HOST", "127.0.0.1")
	defer os.Unsetenv("MCP_HTTP_PORT")
	defer os.Unsetenv("MCP_HTTP_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.HTTP.Port != "9090" {
		t.Errorf("Expected HTTP port '9090', got %q", cfg.HTTP.Port)
	}
	if cfg.HTTP.Host != "127.0.0.1" {
		t.Errorf("Expected HTTP host '127.0.0.1', got %q", cfg.HTTP.Host)
	}
}

func TestHTTPConfigPortOnlyFromEnv(t *testing.T) {
	os.Setenv("TALLY_HOST", "localhost")
	os.Setenv("TALLY_PORT", "9900")
	os.Setenv("TALLY_COMPANY", "DemoCompany")
	os.Setenv("MCP_HTTP_PORT", "8080")
	defer os.Unsetenv("MCP_HTTP_PORT")
	os.Unsetenv("MCP_HTTP_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.HTTP.Port != "8080" {
		t.Errorf("Expected HTTP port '8080', got %q", cfg.HTTP.Port)
	}
	if cfg.HTTP.Host != "0.0.0.0" {
		t.Errorf("Expected default HTTP host '0.0.0.0' when MCP_HTTP_HOST unset, got %q", cfg.HTTP.Host)
	}
}
