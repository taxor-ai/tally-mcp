package logger

import (
	"testing"
)

func TestLoggerInit(t *testing.T) {
	log, err := New("info", "")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if log == nil {
		t.Error("Expected logger, got nil")
	}
}

func TestLoggerWithFile(t *testing.T) {
	log, err := New("debug", "/tmp/test.log")
	if err != nil {
		t.Fatalf("New() with file failed: %v", err)
	}
	if log == nil {
		t.Error("Expected logger, got nil")
	}
	log.Info("Test message", "key", "value")
}
