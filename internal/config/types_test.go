package config

import (
	"testing"
)

func TestServerDefaults(t *testing.T) {
	defaults := ServerDefaults()

	if defaults.FallbackHost != "storage.googleapis.com" {
		t.Errorf("FallbackHost default = %q, expected %q", defaults.FallbackHost, "storage.googleapis.com")
	}
	if defaults.MultiplexPerConnection != 50 {
		t.Errorf("MultiplexPerConnection default = %d, expected 50", defaults.MultiplexPerConnection)
	}
	if defaults.LogLevel != "warn" {
		t.Errorf("LogLevel default = %q, expected %q", defaults.LogLevel, "warn")
	}

	// These should be zero values
	if defaults.ListenPort != 0 {
		t.Errorf("ListenPort should be 0, got %d", defaults.ListenPort)
	}
	if defaults.MTProtoPort != 0 {
		t.Errorf("MTProtoPort should be 0, got %d", defaults.MTProtoPort)
	}
	if defaults.VMESSPort != 0 {
		t.Errorf("VMESSPort should be 0, got %d", defaults.VMESSPort)
	}
}

func TestFreenetDefaults(t *testing.T) {
	defaults := FreenetDefaults()

	if defaults.MaxConnections != 30 {
		t.Errorf("MaxConnections default = %d, expected 30", defaults.MaxConnections)
	}
	if defaults.LogLevel != "warn" {
		t.Errorf("LogLevel default = %q, expected %q", defaults.LogLevel, "warn")
	}

	// These should be empty
	if defaults.ServerIP != "" {
		t.Errorf("ServerIP should be empty")
	}
	if defaults.FreeInterface != "" {
		t.Errorf("FreeInterface should be empty")
	}
}

func TestServerConfigValidate(t *testing.T) {
	validConfig := ServerConfig{
		ListenPort:             2083,
		MTProtoPort:            8443,
		VMESSPort:              8081,
		MTProtoSecret:          "0123456789abcdef0123456789abcdef",
		VMESSUuid:              "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FallbackHost:           "storage.googleapis.com",
		MultiplexPerConnection: 50,
		LogLevel:               "warn",
	}

	tests := []struct {
		name      string
		modify    func(*ServerConfig)
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid config",
			modify:    func(c *ServerConfig) {},
			expectErr: false,
		},
		{
			name:      "missing listen_port",
			modify:    func(c *ServerConfig) { c.ListenPort = 0 },
			expectErr: true,
			errMsg:    "listen_port is required",
		},
		{
			name:      "missing mtproto_port",
			modify:    func(c *ServerConfig) { c.MTProtoPort = 0 },
			expectErr: true,
			errMsg:    "mtproto_port is required",
		},
		{
			name:      "missing vmess_port",
			modify:    func(c *ServerConfig) { c.VMESSPort = 0 },
			expectErr: true,
			errMsg:    "vmess_port is required",
		},
		{
			name:      "missing mtproto_secret",
			modify:    func(c *ServerConfig) { c.MTProtoSecret = "" },
			expectErr: true,
			errMsg:    "mtproto_secret is required",
		},
		{
			name:      "short mtproto_secret",
			modify:    func(c *ServerConfig) { c.MTProtoSecret = "0123456789abcdef" },
			expectErr: true,
			errMsg:    "mtproto_secret must be 32 hex characters",
		},
		{
			name:      "long mtproto_secret",
			modify:    func(c *ServerConfig) { c.MTProtoSecret = "0123456789abcdef0123456789abcdef00" },
			expectErr: true,
			errMsg:    "mtproto_secret must be 32 hex characters",
		},
		{
			name:      "missing vmess_uuid",
			modify:    func(c *ServerConfig) { c.VMESSUuid = "" },
			expectErr: true,
			errMsg:    "vmess_uuid is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig
			tt.modify(&config)

			err := config.Validate()
			if tt.expectErr {
				if err == nil {
					t.Error("Validate() should have returned error")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Error = %q, expected %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() returned unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFreenetConfigValidate(t *testing.T) {
	validConfig := FreenetConfig{
		ServerIP:            "192.168.1.100",
		ServerPort:          2083,
		VMESSPort:           8081,
		VMESSUuid:           "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FreeInterface:       "eth0",
		RestrictedInterface: "wlan0",
		MaxConnections:      30,
		LogLevel:            "warn",
	}

	tests := []struct {
		name      string
		modify    func(*FreenetConfig)
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid config",
			modify:    func(c *FreenetConfig) {},
			expectErr: false,
		},
		{
			name:      "missing server_ip",
			modify:    func(c *FreenetConfig) { c.ServerIP = "" },
			expectErr: true,
			errMsg:    "server_ip is required",
		},
		{
			name:      "missing server_port",
			modify:    func(c *FreenetConfig) { c.ServerPort = 0 },
			expectErr: true,
			errMsg:    "server_port is required",
		},
		{
			name:      "missing vmess_port",
			modify:    func(c *FreenetConfig) { c.VMESSPort = 0 },
			expectErr: true,
			errMsg:    "vmess_port is required",
		},
		{
			name:      "missing vmess_uuid",
			modify:    func(c *FreenetConfig) { c.VMESSUuid = "" },
			expectErr: true,
			errMsg:    "vmess_uuid is required",
		},
		{
			name:      "missing free_interface",
			modify:    func(c *FreenetConfig) { c.FreeInterface = "" },
			expectErr: true,
			errMsg:    "free_interface is required",
		},
		{
			name:      "missing restricted_interface",
			modify:    func(c *FreenetConfig) { c.RestrictedInterface = "" },
			expectErr: true,
			errMsg:    "restricted_interface is required",
		},
		{
			name:      "same interfaces",
			modify:    func(c *FreenetConfig) { c.RestrictedInterface = c.FreeInterface },
			expectErr: true,
			errMsg:    "free_interface and restricted_interface must be different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig
			tt.modify(&config)

			err := config.Validate()
			if tt.expectErr {
				if err == nil {
					t.Error("Validate() should have returned error")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Error = %q, expected %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() returned unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	msg := "test error message"
	err := NewConfigError(msg)

	if err.Error() != msg {
		t.Errorf("Error() = %q, expected %q", err.Error(), msg)
	}

	if err.Message != msg {
		t.Errorf("Message = %q, expected %q", err.Message, msg)
	}
}
