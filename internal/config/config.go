package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration loaded from environment variables.
// This provides a centralized way to manage all configuration settings.
type Config struct {
	// EdgePort is the HTTP port where the edge router listens for incoming requests.
	// Default: 4566 (matching LocalStack's default port pattern)
	EdgePort int

	// DataDir is the base directory where service data (blobs, state, etc.) is stored.
	// Default: ./data
	DataDir string

	// EnabledServices is a comma-separated list of service names to enable at startup.
	// Example: "blob,queue,keyvault"
	// Default: "blob"
	EnabledServices []string

	// LogLevel controls the verbosity of logging (debug, info, warn, error).
	// Default: "info"
	LogLevel string
}

// Load creates a Config instance by reading environment variables.
// Missing values are replaced with sensible defaults.
func Load() *Config {
	cfg := &Config{
		EdgePort:        4566,
		DataDir:         "./data",
		EnabledServices: []string{"blob"},
		LogLevel:        "info",
	}

	// Load EDGE_PORT
	if portStr := os.Getenv("EDGE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			cfg.EdgePort = port
		}
	}

	// Load DATA_DIR
	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		cfg.DataDir = dataDir
	}

	// Load ENABLED_SERVICES
	if servicesStr := os.Getenv("ENABLED_SERVICES"); servicesStr != "" {
		services := strings.Split(servicesStr, ",")
		enabled := make([]string, 0, len(services))
		for _, s := range services {
			s = strings.TrimSpace(s)
			if s != "" {
				enabled = append(enabled, s)
			}
		}
		if len(enabled) > 0 {
			cfg.EnabledServices = enabled
		}
	}

	// Load LOG_LEVEL
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	return cfg
}

// IsServiceEnabled checks if a given service name is in the EnabledServices list.
func (c *Config) IsServiceEnabled(serviceName string) bool {
	for _, s := range c.EnabledServices {
		if s == serviceName {
			return true
		}
	}
	return false
}

// Validate performs basic validation on the configuration.
// Returns an error if any invalid settings are detected.
func (c *Config) Validate() error {
	if c.EdgePort <= 0 || c.EdgePort >= 65536 {
		return fmt.Errorf("invalid EDGE_PORT: %d (must be 1-65535)", c.EdgePort)
	}
	if c.DataDir == "" {
		return fmt.Errorf("DATA_DIR cannot be empty")
	}
	return nil
}

