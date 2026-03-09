package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerConfigValidation(t *testing.T) {
	cfg := &ServerConfig{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty config")
	}

	cfg = &ServerConfig{
		SocksPort:  1080,
		TunnelPort: 2083,
		UUID:       "test-uuid",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClientConfigValidation(t *testing.T) {
	cfg := &ClientConfig{}
	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty config")
	}

	cfg = &ClientConfig{
		ServerIP:            "10.0.0.1",
		TunnelPort:          2083,
		UUID:                "test-uuid",
		FreeInterface:       "eth0",
		RestrictedInterface: "eth1",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveAndLoadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	original := &ServerConfig{
		SocksPort:  1080,
		TunnelPort: 2083,
		UUID:       "test-uuid-123",
	}

	if err := SaveJSON(path, original); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	var loaded ServerConfig
	if err := LoadJSON(path, &loaded); err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if loaded.SocksPort != original.SocksPort {
		t.Errorf("SocksPort: expected %d, got %d", original.SocksPort, loaded.SocksPort)
	}
	if loaded.UUID != original.UUID {
		t.Errorf("UUID: expected %s, got %s", original.UUID, loaded.UUID)
	}
}
