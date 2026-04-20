//go:build integration

package main

import (
	"bufio"
	"os"
	"strings"
)

func init() {
	// Load environment from .env.test before running integration tests
	loadEnvFile(".env.test")
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// Try from project root (when running tests from subdirectory)
		for _, path := range []string{".env.test", "../../.env.test"} {
			if f, err := os.Open(path); err == nil {
				file = f
				break
			}
		}
		if file == nil {
			return
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Only set if not already set in environment (env vars take precedence)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
