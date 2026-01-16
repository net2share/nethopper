package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nethopper/nethopper/internal/platform"
)

// Manager handles configuration file operations
type Manager struct {
	paths platform.PathProvider
}

// NewManager creates a new configuration manager
func NewManager(paths platform.PathProvider) *Manager {
	return &Manager{paths: paths}
}

// GetConfigPath returns the config path for the given mode
func (m *Manager) GetConfigPath(systemMode bool) string {
	return m.paths.ConfigPath(systemMode)
}

// ConfigExists checks if a configuration file exists
func (m *Manager) ConfigExists(systemMode bool) bool {
	path := m.paths.ConfigPath(systemMode)
	_, err := os.Stat(path)
	return err == nil
}

// WriteServerConfig generates and writes server configuration
func (m *Manager) WriteServerConfig(config *ServerConfig) error {
	data, err := GenerateServerConfig(config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	return m.writeConfig(true, data)
}

// WriteFreenetConfig generates and writes freenet configuration
func (m *Manager) WriteFreenetConfig(config *FreenetConfig) error {
	data, err := GenerateFreenetConfig(config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	return m.writeConfig(false, data)
}

// WriteCustomConfig writes a custom configuration
func (m *Manager) WriteCustomConfig(systemMode bool, data []byte) error {
	// Validate JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		return fmt.Errorf("invalid JSON configuration: %w", err)
	}

	return m.writeConfig(systemMode, data)
}

// writeConfig writes configuration data to file
func (m *Manager) writeConfig(systemMode bool, data []byte) error {
	path := m.paths.ConfigPath(systemMode)
	dir := filepath.Dir(path)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	// Write config file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}

// ReadRawConfig reads the raw configuration file
func (m *Manager) ReadRawConfig(systemMode bool) ([]byte, error) {
	path := m.paths.ConfigPath(systemMode)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", path, err)
	}
	return data, nil
}

// LoadServerConfig loads and parses server configuration
func (m *Manager) LoadServerConfig() (*ServerConfig, error) {
	data, err := m.ReadRawConfig(true)
	if err != nil {
		return nil, err
	}

	return ParseServerConfig(data)
}

// LoadFreenetConfig loads and parses freenet configuration
func (m *Manager) LoadFreenetConfig() (*FreenetConfig, error) {
	data, err := m.ReadRawConfig(false)
	if err != nil {
		return nil, err
	}

	return ParseFreenetConfig(data)
}

// RemoveConfig removes the configuration file
func (m *Manager) RemoveConfig(systemMode bool) error {
	path := m.paths.ConfigPath(systemMode)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config at %s: %w", path, err)
	}
	return nil
}

// ParseServerConfig parses raw JSON into ServerConfig
// Extracts the configurable fields from sing-box JSON format
func ParseServerConfig(data []byte) (*ServerConfig, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	config := ServerDefaults()

	// Extract log level
	if log, ok := raw["log"].(map[string]interface{}); ok {
		if disabled, ok := log["disabled"].(bool); ok && disabled {
			config.LogLevel = "disabled"
		} else if level, ok := log["level"].(string); ok {
			config.LogLevel = level
		}
	}

	// Extract endpoint listen port
	if endpoints, ok := raw["endpoints"].([]interface{}); ok && len(endpoints) > 0 {
		if ep, ok := endpoints[0].(map[string]interface{}); ok {
			if port, ok := ep["listen_port"].(float64); ok {
				config.ListenPort = uint16(port)
			}
		}
	}

	// Extract inbound settings
	if inbounds, ok := raw["inbounds"].([]interface{}); ok {
		for _, ib := range inbounds {
			if inbound, ok := ib.(map[string]interface{}); ok {
				if inbound["type"] == "mtproto" {
					if port, ok := inbound["listen_port"].(float64); ok {
						config.MTProtoPort = uint16(port)
					}
					if multiplex, ok := inbound["multiplex_per_connection"].(float64); ok {
						config.MultiplexPerConnection = int(multiplex)
					}
					if fallback, ok := inbound["fallback_host"].(string); ok {
						config.FallbackHost = fallback
					}
					if users, ok := inbound["users"].([]interface{}); ok && len(users) > 0 {
						if user, ok := users[0].(map[string]interface{}); ok {
							if secret, ok := user["secret"].(string); ok {
								config.MTProtoSecret = secret
							}
						}
					}
				}
			}
		}
	}

	// Extract outbound VMess settings
	if outbounds, ok := raw["outbounds"].([]interface{}); ok {
		for _, ob := range outbounds {
			if outbound, ok := ob.(map[string]interface{}); ok {
				if outbound["type"] == "vmess" {
					if uuid, ok := outbound["uuid"].(string); ok {
						config.VMESSUuid = uuid
					}
					if port, ok := outbound["server_port"].(float64); ok {
						config.VMESSPort = uint16(port)
					}
				}
			}
		}
	}

	return &config, nil
}

// ParseFreenetConfig parses raw JSON into FreenetConfig
func ParseFreenetConfig(data []byte) (*FreenetConfig, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	config := FreenetDefaults()

	// Extract log level
	if log, ok := raw["log"].(map[string]interface{}); ok {
		if disabled, ok := log["disabled"].(bool); ok && disabled {
			config.LogLevel = "disabled"
		} else if level, ok := log["level"].(string); ok {
			config.LogLevel = level
		}
	}

	// Extract endpoint settings
	if endpoints, ok := raw["endpoints"].([]interface{}); ok && len(endpoints) > 0 {
		if ep, ok := endpoints[0].(map[string]interface{}); ok {
			if server, ok := ep["server"].(string); ok {
				config.ServerIP = server
			}
			if port, ok := ep["server_port"].(float64); ok {
				config.ServerPort = uint16(port)
			}
			if multiplex, ok := ep["multiplex"].(map[string]interface{}); ok {
				if maxConn, ok := multiplex["max_connections"].(float64); ok && int(maxConn) > 0 {
					config.MaxConnections = int(maxConn)
				}
			}
		}
	}

	// Extract inbound VMess settings
	if inbounds, ok := raw["inbounds"].([]interface{}); ok {
		for _, ib := range inbounds {
			if inbound, ok := ib.(map[string]interface{}); ok {
				if inbound["type"] == "vmess" {
					if port, ok := inbound["listen_port"].(float64); ok {
						config.VMESSPort = uint16(port)
					}
					if users, ok := inbound["users"].([]interface{}); ok && len(users) > 0 {
						if user, ok := users[0].(map[string]interface{}); ok {
							if uuid, ok := user["uuid"].(string); ok {
								config.VMESSUuid = uuid
							}
						}
					}
				}
			}
		}
	}

	// Extract outbound interface settings
	if outbounds, ok := raw["outbounds"].([]interface{}); ok {
		for _, ob := range outbounds {
			if outbound, ok := ob.(map[string]interface{}); ok {
				if outbound["type"] == "direct" {
					tag, _ := outbound["tag"].(string)
					iface, _ := outbound["bind_interface"].(string)
					if tag == "free" {
						config.FreeInterface = iface
					} else if tag == "filternet" {
						config.RestrictedInterface = iface
					}
				}
			}
		}
	}

	return &config, nil
}
