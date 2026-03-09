package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ServerConfig holds server-side configuration.
type ServerConfig struct {
	SocksPort  int    `json:"socks_port"`
	TunnelPort int    `json:"tunnel_port"`
	UUID       string `json:"uuid"`
}

// ClientConfig holds client-side configuration.
type ClientConfig struct {
	ServerIP            string `json:"server_ip"`
	TunnelPort          int    `json:"tunnel_port"`
	SocksPort           int    `json:"socks_port"`
	UUID                string `json:"uuid"`
	FreeInterface       string `json:"free_interface"`
	RestrictedInterface string `json:"restricted_interface"`
}

// Validate validates the server config.
func (c *ServerConfig) Validate() error {
	if c.SocksPort < 1 || c.SocksPort > 65535 {
		return fmt.Errorf("invalid socks port: %d", c.SocksPort)
	}
	if c.TunnelPort < 1 || c.TunnelPort > 65535 {
		return fmt.Errorf("invalid tunnel port: %d", c.TunnelPort)
	}
	if c.UUID == "" {
		return fmt.Errorf("UUID is required")
	}
	return nil
}

// Validate validates the client config.
func (c *ClientConfig) Validate() error {
	if c.ServerIP == "" {
		return fmt.Errorf("server IP is required")
	}
	if c.TunnelPort < 1 || c.TunnelPort > 65535 {
		return fmt.Errorf("invalid tunnel port: %d", c.TunnelPort)
	}
	if c.UUID == "" {
		return fmt.Errorf("UUID is required")
	}
	if c.FreeInterface == "" {
		return fmt.Errorf("free interface is required")
	}
	if c.RestrictedInterface == "" {
		return fmt.Errorf("restricted interface is required")
	}
	if c.FreeInterface == c.RestrictedInterface {
		return fmt.Errorf("free and restricted interfaces must be different")
	}
	return nil
}

// SaveJSON saves a config struct to a JSON file.
func SaveJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0640); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// LoadJSON loads a config struct from a JSON file.
func LoadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	return nil
}
