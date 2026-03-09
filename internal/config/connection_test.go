package config

import (
	"strings"
	"testing"
)

func TestConnectionStringRoundTrip(t *testing.T) {
	cs := &ConnectionString{
		Version:    1,
		ServerIP:   "192.168.1.100",
		TunnelPort: 2083,
		SocksPort:  1080,
		UUID:       "test-uuid-1234",
	}

	encoded, err := cs.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.HasPrefix(encoded, ConnectionStringPrefix) {
		t.Fatalf("expected prefix %s, got %s", ConnectionStringPrefix, encoded)
	}

	decoded, err := DecodeConnectionString(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.ServerIP != cs.ServerIP {
		t.Errorf("ServerIP: expected %s, got %s", cs.ServerIP, decoded.ServerIP)
	}
	if decoded.TunnelPort != cs.TunnelPort {
		t.Errorf("TunnelPort: expected %d, got %d", cs.TunnelPort, decoded.TunnelPort)
	}
	if decoded.SocksPort != cs.SocksPort {
		t.Errorf("SocksPort: expected %d, got %d", cs.SocksPort, decoded.SocksPort)
	}
	if decoded.UUID != cs.UUID {
		t.Errorf("UUID: expected %s, got %s", cs.UUID, decoded.UUID)
	}
}

func TestIsConnectionString(t *testing.T) {
	if !IsConnectionString("nh://abc123") {
		t.Error("should be a connection string")
	}
	if IsConnectionString("http://example.com") {
		t.Error("should not be a connection string")
	}
}

func TestDecodeInvalidConnectionString(t *testing.T) {
	tests := []string{
		"invalid",
		"nh://",
		"nh://!!!invalid-base64!!!",
	}
	for _, tc := range tests {
		_, err := DecodeConnectionString(tc)
		if err == nil {
			t.Errorf("expected error for %q", tc)
		}
	}
}

func TestConnectionStringValidation(t *testing.T) {
	cs := &ConnectionString{}
	if err := cs.Validate(); err == nil {
		t.Error("expected validation error for empty connection string")
	}

	cs.ServerIP = "1.2.3.4"
	cs.TunnelPort = 2083
	cs.UUID = "test-uuid"
	if err := cs.Validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestToClientConfig(t *testing.T) {
	cs := &ConnectionString{
		ServerIP:   "10.0.0.1",
		TunnelPort: 2083,
		SocksPort:  1080,
		UUID:       "uuid-123",
	}

	cfg := cs.ToClientConfig("eth0", "eth1")
	if cfg.ServerIP != "10.0.0.1" {
		t.Errorf("unexpected ServerIP: %s", cfg.ServerIP)
	}
	if cfg.FreeInterface != "eth0" {
		t.Errorf("unexpected FreeInterface: %s", cfg.FreeInterface)
	}
	if cfg.RestrictedInterface != "eth1" {
		t.Errorf("unexpected RestrictedInterface: %s", cfg.RestrictedInterface)
	}
}
