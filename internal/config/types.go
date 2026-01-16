package config

// ServerConfig represents all configurable server parameters
type ServerConfig struct {
	ListenPort             uint16 `json:"listen_port"`               // Reverse endpoint port
	MTProtoPort            uint16 `json:"mtproto_port"`              // MTProto inbound port
	VMESSPort              uint16 `json:"vmess_port"`                // Outbound VMess port
	MTProtoSecret          string `json:"mtproto_secret"`            // 32-char hex secret
	VMESSUuid              string `json:"vmess_uuid"`                // VMess UUID
	FallbackHost           string `json:"fallback_host"`             // Domain for TLS camouflage
	MultiplexPerConnection int    `json:"multiplex_per_connection"`  // Streams per connection
	LogLevel               string `json:"log_level"`                 // Log level
}

// FreenetConfig represents all configurable freenet parameters
type FreenetConfig struct {
	ServerIP            string `json:"server_ip"`            // Server IP in restricted network
	ServerPort          uint16 `json:"server_port"`          // Reverse endpoint port
	VMESSPort           uint16 `json:"vmess_port"`           // Internal VMess port
	VMESSUuid           string `json:"vmess_uuid"`           // VMess UUID
	FreeInterface       string `json:"free_interface"`       // Interface with free internet
	RestrictedInterface string `json:"restricted_interface"` // Interface on restricted network
	MaxConnections      int    `json:"max_connections"`      // Multiplex max connections
	LogLevel            string `json:"log_level"`            // Log level
}

// ServerDefaults returns default values for server config
func ServerDefaults() ServerConfig {
	return ServerConfig{
		FallbackHost:           "storage.googleapis.com",
		MultiplexPerConnection: 50,
		LogLevel:               "warn",
	}
}

// FreenetDefaults returns default values for freenet config
func FreenetDefaults() FreenetConfig {
	return FreenetConfig{
		MaxConnections: 30,
		LogLevel:       "warn",
	}
}

// Validate validates the server configuration and applies defaults
func (c *ServerConfig) Validate() error {
	if c.ListenPort == 0 {
		return NewConfigError("listen_port is required")
	}
	if c.MTProtoPort == 0 {
		return NewConfigError("mtproto_port is required")
	}
	if c.VMESSPort == 0 {
		return NewConfigError("vmess_port is required")
	}
	if c.MTProtoSecret == "" {
		return NewConfigError("mtproto_secret is required")
	}
	if len(c.MTProtoSecret) != 32 {
		return NewConfigError("mtproto_secret must be 32 hex characters")
	}
	if c.VMESSUuid == "" {
		return NewConfigError("vmess_uuid is required")
	}

	// Apply defaults for optional fields
	if c.FallbackHost == "" {
		c.FallbackHost = "storage.googleapis.com"
	}
	if c.MultiplexPerConnection <= 0 {
		c.MultiplexPerConnection = 50
	}
	if c.LogLevel == "" {
		c.LogLevel = "warn"
	}

	return nil
}

// Validate validates the freenet configuration and applies defaults
func (c *FreenetConfig) Validate() error {
	if c.ServerIP == "" {
		return NewConfigError("server_ip is required")
	}
	if c.ServerPort == 0 {
		return NewConfigError("server_port is required")
	}
	if c.VMESSPort == 0 {
		return NewConfigError("vmess_port is required")
	}
	if c.VMESSUuid == "" {
		return NewConfigError("vmess_uuid is required")
	}
	if c.FreeInterface == "" {
		return NewConfigError("free_interface is required")
	}
	if c.RestrictedInterface == "" {
		return NewConfigError("restricted_interface is required")
	}
	if c.FreeInterface == c.RestrictedInterface {
		return NewConfigError("free_interface and restricted_interface must be different")
	}

	// Apply defaults for optional fields
	if c.MaxConnections <= 0 {
		c.MaxConnections = 30
	}
	if c.LogLevel == "" {
		c.LogLevel = "warn"
	}

	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

// NewConfigError creates a new ConfigError
func NewConfigError(msg string) *ConfigError {
	return &ConfigError{Message: msg}
}
