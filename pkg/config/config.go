package config

import (
	"fmt"
	"os"
	"strconv"
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
