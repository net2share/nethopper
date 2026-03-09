package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ConnectionStringPrefix  = "nh://"
	ConnectionStringVersion = 1
)

// ConnectionString represents the compact format for sharing server config.
type ConnectionString struct {
	Version    uint8  `json:"v"`
	ServerIP   string `json:"s"`
	TunnelPort int    `json:"p"`
	SocksPort  int    `json:"sp"`
	UUID       string `json:"u"`
}

// NewConnectionString creates a ConnectionString from server config and IP.
func NewConnectionString(serverIP string, cfg *ServerConfig) *ConnectionString {
	return &ConnectionString{
		Version:    ConnectionStringVersion,
		ServerIP:   serverIP,
		TunnelPort: cfg.TunnelPort,
		SocksPort:  cfg.SocksPort,
		UUID:       cfg.UUID,
	}
}

// Encode encodes the ConnectionString to nh:// format.
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

// Validate validates the ConnectionString.
func (cs *ConnectionString) Validate() error {
	if cs.ServerIP == "" {
		return fmt.Errorf("server IP is required")
	}
	if cs.TunnelPort == 0 {
		return fmt.Errorf("tunnel port is required")
	}
	if cs.UUID == "" {
		return fmt.Errorf("UUID is required")
	}
	return nil
}

// DecodeConnectionString decodes an nh:// connection string.
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

// IsConnectionString checks if a string looks like a connection string.
func IsConnectionString(s string) bool {
	return strings.HasPrefix(s, ConnectionStringPrefix)
}

// ToClientConfig converts the ConnectionString to a partial ClientConfig.
func (cs *ConnectionString) ToClientConfig(freeInterface, restrictedInterface string) *ClientConfig {
	return &ClientConfig{
		ServerIP:            cs.ServerIP,
		TunnelPort:          cs.TunnelPort,
		SocksPort:           cs.SocksPort,
		UUID:                cs.UUID,
		FreeInterface:       freeInterface,
		RestrictedInterface: restrictedInterface,
	}
}

// String returns a human-readable representation.
func (cs *ConnectionString) String() string {
	return fmt.Sprintf("Server: %s, Tunnel: %d, SOCKS5: %d", cs.ServerIP, cs.TunnelPort, cs.SocksPort)
}
