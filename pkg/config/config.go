package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	MaxPlayers   int           `json:"max_players"`
	TickRate     int           `json:"tick_rate"` // Updates per second
	WorldSize    int           `json:"world_size"` // Size of world in chunks
	ChunkSize    int           `json:"chunk_size"` // Size of each chunk
	ViewDistance int           `json:"view_distance"` // How many chunks to send to clients
	Timeout      time.Duration `json:"timeout"` // Connection timeout
	EnableTLS    bool          `json:"enable_tls"` // Enable TLS encryption
	CertFile     string        `json:"cert_file"` // TLS certificate file
	KeyFile      string        `json:"key_file"` // TLS private key file
	LogLevel     string        `json:"log_level"` // Debug, Info, Warn, Error
	Environment  string        `json:"environment"` // development, production, staging
}

// DefaultConfig returns a default server configuration
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Host:         "0.0.0.0",
		Port:         8080,
		MaxPlayers:   32,
		TickRate:     20,
		WorldSize:    10,
		ChunkSize:    32,
		ViewDistance: 5,
		Timeout:      30 * time.Second,
		EnableTLS:    false,
		LogLevel:     "info",
		Environment:  "development",
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config ServerConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to decode config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return DefaultConfig(), fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config ServerConfig, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *ServerConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.MaxPlayers < 1 {
		return fmt.Errorf("max_players must be at least 1")
	}

	if c.MaxPlayers > 1000 {
		return fmt.Errorf("max_players seems too high (>1000)")
	}

	if c.TickRate < 1 {
		return fmt.Errorf("tick_rate must be at least 1")
	}

	if c.TickRate > 200 {
		return fmt.Errorf("tick_rate seems too high (>200)")
	}

	if c.ChunkSize < 8 {
		return fmt.Errorf("chunk_size must be at least 8")
	}

	if c.ViewDistance < 1 {
		return fmt.Errorf("view_distance must be at least 1")
	}

	if c.EnableTLS {
		if c.CertFile == "" {
			return fmt.Errorf("cert_file is required when TLS is enabled")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("key_file is required when TLS is enabled")
		}
		// Check if files exist
		if _, err := os.Stat(c.CertFile); err != nil {
			return fmt.Errorf("cert_file does not exist: %w", err)
		}
		if _, err := os.Stat(c.KeyFile); err != nil {
			return fmt.Errorf("key_file does not exist: %w", err)
		}
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log_level must be one of: debug, info, warn, error")
	}

	switch c.Environment {
	case "development", "production", "staging":
	default:
		return fmt.Errorf("environment must be one of: development, production, staging")
	}

	return nil
}