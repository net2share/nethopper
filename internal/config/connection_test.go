package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewConnectionString(t *testing.T) {
	cs := NewConnectionString("192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23")

	if cs.Version != ConnectionStringVersion {
		t.Errorf("Version = %d, expected %d", cs.Version, ConnectionStringVersion)
	}
	if cs.ServerIP != "192.168.1.100" {
		t.Errorf("ServerIP = %q, expected %q", cs.ServerIP, "192.168.1.100")
	}
	if cs.Port != 2083 {
		t.Errorf("Port = %d, expected %d", cs.Port, 2083)
	}
	if cs.VMESSPort != 8081 {
		t.Errorf("VMESSPort = %d, expected %d", cs.VMESSPort, 8081)
	}
	if cs.UUID != "103b0aae-3384-4d23-9f5b-2d15be377a23" {
		t.Errorf("UUID = %q, expected %q", cs.UUID, "103b0aae-3384-4d23-9f5b-2d15be377a23")
	}
}

func TestFromServerConfig(t *testing.T) {
	serverConfig := &ServerConfig{
		ListenPort: 2083,
		VMESSPort:  8081,
		VMESSUuid:  "103b0aae-3384-4d23-9f5b-2d15be377a23",
	}

	cs := FromServerConfig("10.0.0.1", serverConfig)

	if cs.ServerIP != "10.0.0.1" {
		t.Errorf("ServerIP = %q, expected %q", cs.ServerIP, "10.0.0.1")
	}
	if cs.Port != serverConfig.ListenPort {
		t.Errorf("Port = %d, expected %d", cs.Port, serverConfig.ListenPort)
	}
	if cs.VMESSPort != serverConfig.VMESSPort {
		t.Errorf("VMESSPort = %d, expected %d", cs.VMESSPort, serverConfig.VMESSPort)
	}
	if cs.UUID != serverConfig.VMESSUuid {
		t.Errorf("UUID = %q, expected %q", cs.UUID, serverConfig.VMESSUuid)
	}
}

func TestConnectionStringEncode(t *testing.T) {
	cs := NewConnectionString("192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23")

	encoded, err := cs.Encode()
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// Should have the correct prefix
	if !strings.HasPrefix(encoded, ConnectionStringPrefix) {
		t.Errorf("Encoded string does not start with %q: %s", ConnectionStringPrefix, encoded)
	}

	// Should be decodable base64 after removing prefix
	base64Part := strings.TrimPrefix(encoded, ConnectionStringPrefix)
	decoded, err := base64.URLEncoding.DecodeString(base64Part)
	if err != nil {
		t.Fatalf("Base64 decode failed: %v", err)
	}

	// Should be valid JSON
	var result ConnectionString
	if err := json.Unmarshal(decoded, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Should match original values
	if result.ServerIP != cs.ServerIP {
		t.Errorf("Decoded ServerIP = %q, expected %q", result.ServerIP, cs.ServerIP)
	}
	if result.Port != cs.Port {
		t.Errorf("Decoded Port = %d, expected %d", result.Port, cs.Port)
	}
}

func TestConnectionStringValidate(t *testing.T) {
	tests := []struct {
		name      string
		cs        ConnectionString
		expectErr bool
	}{
		{
			name: "valid connection string",
			cs: ConnectionString{
				Version:   1,
				ServerIP:  "192.168.1.100",
				Port:      2083,
				VMESSPort: 8081,
				UUID:      "103b0aae-3384-4d23-9f5b-2d15be377a23",
			},
			expectErr: false,
		},
		{
			name: "missing version",
			cs: ConnectionString{
				Version:   0,
				ServerIP:  "192.168.1.100",
				Port:      2083,
				VMESSPort: 8081,
				UUID:      "103b0aae-3384-4d23-9f5b-2d15be377a23",
			},
			expectErr: true,
		},
		{
			name: "missing server IP",
			cs: ConnectionString{
				Version:   1,
				ServerIP:  "",
				Port:      2083,
				VMESSPort: 8081,
				UUID:      "103b0aae-3384-4d23-9f5b-2d15be377a23",
			},
			expectErr: true,
		},
		{
			name: "missing port",
			cs: ConnectionString{
				Version:   1,
				ServerIP:  "192.168.1.100",
				Port:      0,
				VMESSPort: 8081,
				UUID:      "103b0aae-3384-4d23-9f5b-2d15be377a23",
			},
			expectErr: true,
		},
		{
			name: "missing vmess port",
			cs: ConnectionString{
				Version:   1,
				ServerIP:  "192.168.1.100",
				Port:      2083,
				VMESSPort: 0,
				UUID:      "103b0aae-3384-4d23-9f5b-2d15be377a23",
			},
			expectErr: true,
		},
		{
			name: "missing UUID",
			cs: ConnectionString{
				Version:   1,
				ServerIP:  "192.168.1.100",
				Port:      2083,
				VMESSPort: 8081,
				UUID:      "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cs.Validate()
			if tt.expectErr && err == nil {
				t.Error("Validate() should have returned error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Validate() returned unexpected error: %v", err)
			}
		})
	}
}

func TestDecodeConnectionString(t *testing.T) {
	// Create and encode a connection string
	original := NewConnectionString("192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23")
	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	// Decode it
	decoded, err := DecodeConnectionString(encoded)
	if err != nil {
		t.Fatalf("DecodeConnectionString() returned error: %v", err)
	}

	// Verify all fields match
	if decoded.Version != original.Version {
		t.Errorf("Version = %d, expected %d", decoded.Version, original.Version)
	}
	if decoded.ServerIP != original.ServerIP {
		t.Errorf("ServerIP = %q, expected %q", decoded.ServerIP, original.ServerIP)
	}
	if decoded.Port != original.Port {
		t.Errorf("Port = %d, expected %d", decoded.Port, original.Port)
	}
	if decoded.VMESSPort != original.VMESSPort {
		t.Errorf("VMESSPort = %d, expected %d", decoded.VMESSPort, original.VMESSPort)
	}
	if decoded.UUID != original.UUID {
		t.Errorf("UUID = %q, expected %q", decoded.UUID, original.UUID)
	}
}

func TestDecodeConnectionStringErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "missing prefix",
			input:       "eyJ2IjoxLCJzIjoiMTkyLjE2OC4xLjEwMCIsInAiOjIwODMsInZwIjo4MDgxLCJ1IjoiMTAzYjBhYWUtMzM4NC00ZDIzLTlmNWItMmQxNWJlMzc3YTIzIn0=",
			expectError: true,
		},
		{
			name:        "wrong prefix",
			input:       "wrong://eyJ2IjoxLCJzIjoiMTkyLjE2OC4xLjEwMCIsInAiOjIwODMsInZwIjo4MDgxLCJ1IjoiMTAzYjBhYWUtMzM4NC00ZDIzLTlmNWItMmQxNWJlMzc3YTIzIn0=",
			expectError: true,
		},
		{
			name:        "invalid base64",
			input:       "nh://not-valid-base64!!!",
			expectError: true,
		},
		{
			name:        "invalid JSON",
			input:       "nh://" + base64.URLEncoding.EncodeToString([]byte("not json")),
			expectError: true,
		},
		{
			name:        "wrong version",
			input:       "nh://" + base64.URLEncoding.EncodeToString([]byte(`{"v":99,"s":"192.168.1.100","p":2083,"vp":8081,"u":"103b0aae-3384-4d23-9f5b-2d15be377a23"}`)),
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeConnectionString(tt.input)
			if tt.expectError && err == nil {
				t.Error("DecodeConnectionString() should have returned error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("DecodeConnectionString() returned unexpected error: %v", err)
			}
		})
	}
}

func TestIsConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid prefix", "nh://somedata", true},
		{"valid empty data", "nh://", true},
		{"no prefix", "somedata", false},
		{"wrong prefix", "http://example.com", false},
		{"empty string", "", false},
		{"similar prefix", "n2s:", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionString(tt.input)
			if result != tt.expected {
				t.Errorf("IsConnectionString(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToFreenetConfig(t *testing.T) {
	cs := NewConnectionString("192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23")

	// Test with provided values
	fc := cs.ToFreenetConfig("eth0", "wlan0", 50, "debug")

	if fc.ServerIP != cs.ServerIP {
		t.Errorf("ServerIP = %q, expected %q", fc.ServerIP, cs.ServerIP)
	}
	if fc.ServerPort != cs.Port {
		t.Errorf("ServerPort = %d, expected %d", fc.ServerPort, cs.Port)
	}
	if fc.VMESSPort != cs.VMESSPort {
		t.Errorf("VMESSPort = %d, expected %d", fc.VMESSPort, cs.VMESSPort)
	}
	if fc.VMESSUuid != cs.UUID {
		t.Errorf("VMESSUuid = %q, expected %q", fc.VMESSUuid, cs.UUID)
	}
	if fc.FreeInterface != "eth0" {
		t.Errorf("FreeInterface = %q, expected %q", fc.FreeInterface, "eth0")
	}
	if fc.RestrictedInterface != "wlan0" {
		t.Errorf("RestrictedInterface = %q, expected %q", fc.RestrictedInterface, "wlan0")
	}
	if fc.MaxConnections != 50 {
		t.Errorf("MaxConnections = %d, expected %d", fc.MaxConnections, 50)
	}
	if fc.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, expected %q", fc.LogLevel, "debug")
	}
}

func TestToFreenetConfigDefaults(t *testing.T) {
	cs := NewConnectionString("192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23")

	// Test with zero/empty values to trigger defaults
	fc := cs.ToFreenetConfig("eth0", "wlan0", 0, "")

	if fc.MaxConnections != 30 {
		t.Errorf("MaxConnections default = %d, expected 30", fc.MaxConnections)
	}
	if fc.LogLevel != "warn" {
		t.Errorf("LogLevel default = %q, expected %q", fc.LogLevel, "warn")
	}
}

func TestConnectionStringString(t *testing.T) {
	cs := NewConnectionString("192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23")

	str := cs.String()

	// Should contain key information
	if !strings.Contains(str, "192.168.1.100") {
		t.Errorf("String() should contain server IP")
	}
	if !strings.Contains(str, "2083") {
		t.Errorf("String() should contain port")
	}
	if !strings.Contains(str, "8081") {
		t.Errorf("String() should contain VMess port")
	}
}

func TestConnectionStringRoundTrip(t *testing.T) {
	// Test complete round-trip: create -> encode -> decode
	testCases := []struct {
		serverIP  string
		port      uint16
		vmessPort uint16
		uuid      string
	}{
		{"192.168.1.100", 2083, 8081, "103b0aae-3384-4d23-9f5b-2d15be377a23"},
		{"10.0.0.1", 1024, 1025, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"},
		{"172.16.0.1", 49151, 49150, "12345678-1234-1234-1234-123456789012"},
		{"example.com", 8443, 8444, "00000000-0000-0000-0000-000000000000"},
	}

	for _, tc := range testCases {
		original := NewConnectionString(tc.serverIP, tc.port, tc.vmessPort, tc.uuid)
		encoded, err := original.Encode()
		if err != nil {
			t.Errorf("Encode() failed for %+v: %v", tc, err)
			continue
		}

		decoded, err := DecodeConnectionString(encoded)
		if err != nil {
			t.Errorf("DecodeConnectionString() failed for %+v: %v", tc, err)
			continue
		}

		if decoded.ServerIP != tc.serverIP {
			t.Errorf("Round-trip ServerIP mismatch: got %q, want %q", decoded.ServerIP, tc.serverIP)
		}
		if decoded.Port != tc.port {
			t.Errorf("Round-trip Port mismatch: got %d, want %d", decoded.Port, tc.port)
		}
		if decoded.VMESSPort != tc.vmessPort {
			t.Errorf("Round-trip VMESSPort mismatch: got %d, want %d", decoded.VMESSPort, tc.vmessPort)
		}
		if decoded.UUID != tc.uuid {
			t.Errorf("Round-trip UUID mismatch: got %q, want %q", decoded.UUID, tc.uuid)
		}
	}
}

func TestConnectionStringConstants(t *testing.T) {
	if ConnectionStringPrefix != "nh://" {
		t.Errorf("ConnectionStringPrefix = %q, expected %q", ConnectionStringPrefix, "nh://")
	}
	if ConnectionStringVersion != 1 {
		t.Errorf("ConnectionStringVersion = %d, expected 1", ConnectionStringVersion)
	}
}
