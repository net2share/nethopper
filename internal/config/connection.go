package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// ConnectionStringPrefix is the prefix for connection strings
	ConnectionStringPrefix = "nh://"
	// ConnectionStringVersion is the current version of the connection string format
	ConnectionStringVersion = 1
)

// ConnectionString represents the compact format for sharing server config
type ConnectionString struct {
	Version   uint8  `json:"v"`  // Protocol version
	ServerIP  string `json:"s"`  // Server IP or hostname
	Port      uint16 `json:"p"`  // Reverse endpoint port
	VMESSPort uint16 `json:"vp"` // VMess port
	UUID      string `json:"u"`  // VMess UUID
}

// NewConnectionString creates a ConnectionString from server config
func NewConnectionString(serverIP string, port, vmessPort uint16, uuid string) *ConnectionString {
	return &ConnectionString{
		Version:   ConnectionStringVersion,
		ServerIP:  serverIP,
		Port:      port,
		VMESSPort: vmessPort,
		UUID:      uuid,
	}
}

// FromServerConfig creates a ConnectionString from a ServerConfig
// Note: Requires the server IP to be provided separately as it's not in ServerConfig
func FromServerConfig(serverIP string, config *ServerConfig) *ConnectionString {
	return &ConnectionString{
		Version:   ConnectionStringVersion,
		ServerIP:  serverIP,
		Port:      config.ListenPort,
		VMESSPort: config.VMESSPort,
		UUID:      config.VMESSUuid,
	}
}

// Encode encodes the ConnectionString to a string format
func (cs *ConnectionString) Encode() (string, error) {
	if err := cs.Validate(); err != nil {
		return "", err
	}

	data, err := json.Marshal(cs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal connection string: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(data)
	return ConnectionStringPrefix + encoded, nil
}

// Validate validates the ConnectionString
func (cs *ConnectionString) Validate() error {
	if cs.Version == 0 {
		return fmt.Errorf("invalid version: %d", cs.Version)
	}
	if cs.ServerIP == "" {
		return fmt.Errorf("server IP is required")
	}
	if cs.Port == 0 {
		return fmt.Errorf("port is required")
	}
	if cs.VMESSPort == 0 {
		return fmt.Errorf("VMess port is required")
	}
	if cs.UUID == "" {
		return fmt.Errorf("UUID is required")
	}
	return nil
}

// DecodeConnectionString decodes a connection string
func DecodeConnectionString(s string) (*ConnectionString, error) {
	if !strings.HasPrefix(s, ConnectionStringPrefix) {
		return nil, fmt.Errorf("invalid connection string: missing prefix '%s'", ConnectionStringPrefix)
	}

	encoded := strings.TrimPrefix(s, ConnectionStringPrefix)
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: decode failed: %w", err)
	}

	var cs ConnectionString
	if err := json.Unmarshal(data, &cs); err != nil {
		return nil, fmt.Errorf("invalid connection string: parse failed: %w", err)
	}

	if cs.Version != ConnectionStringVersion {
		return nil, fmt.Errorf("unsupported connection string version: %d (expected %d)", cs.Version, ConnectionStringVersion)
	}

	if err := cs.Validate(); err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	return &cs, nil
}

// IsConnectionString checks if a string looks like a connection string
func IsConnectionString(s string) bool {
	return strings.HasPrefix(s, ConnectionStringPrefix)
}

// ToFreenetConfig converts the ConnectionString to a partial FreenetConfig
// Note: Interface names must be provided separately
func (cs *ConnectionString) ToFreenetConfig(freeIface, restrictedIface string, maxConnections int, logLevel string) *FreenetConfig {
	if maxConnections <= 0 {
		maxConnections = 30
	}
	if logLevel == "" {
		logLevel = "warn"
	}

	return &FreenetConfig{
		ServerIP:            cs.ServerIP,
		ServerPort:          cs.Port,
		VMESSPort:           cs.VMESSPort,
		VMESSUuid:           cs.UUID,
		FreeInterface:       freeIface,
		RestrictedInterface: restrictedIface,
		MaxConnections:      maxConnections,
		LogLevel:            logLevel,
	}
}

// String returns a human-readable representation
func (cs *ConnectionString) String() string {
	return fmt.Sprintf("Server: %s:%d, VMess Port: %d", cs.ServerIP, cs.Port, cs.VMESSPort)
}
