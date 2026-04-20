package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Tally  TallyConfig
	Logger LoggerConfig
}

type TallyConfig struct {
	Host    string
	Port    int
	Company string
	Timeout int // seconds
}

type LoggerConfig struct {
	Level string // debug, info, warn, error
	File  string // optional, empty means stdout only
}

func Load() (*Config, error) {
	// Try to load from .env files (in order of precedence)
	envFiles := []string{".env.local", ".env.test", ".env"}
	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			if err := loadEnvFile(envFile); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load %s: %v\n", envFile, err)
			}
			break
		}
	}

	cfg := &Config{
		Tally: TallyConfig{
			Host:    os.Getenv("TALLY_HOST"),
			Company: os.Getenv("TALLY_COMPANY"),
			Timeout: 30,
		},
		Logger: LoggerConfig{
			Level: getEnvOrDefault("TALLY_LOG_LEVEL", "info"),
			File:  os.Getenv("TALLY_LOG_FILE"),
		},
	}

	// Parse port
	portStr := os.Getenv("TALLY_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("TALLY_PORT environment variable is required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TALLY_PORT: %w", err)
	}
	cfg.Tally.Port = port

	// Validate required fields
	if cfg.Tally.Host == "" {
		return nil, fmt.Errorf("TALLY_HOST environment variable is required")
	}
	if cfg.Tally.Company == "" {
		return nil, fmt.Errorf("TALLY_COMPANY environment variable is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
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

	return scanner.Err()
}
