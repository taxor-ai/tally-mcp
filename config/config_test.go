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
