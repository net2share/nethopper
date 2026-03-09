package xray

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/net2share/nethopper/internal/config"
)

func TestGenerateClientConfig(t *testing.T) {
	cfg := &config.ClientConfig{
		ServerIP:            "192.168.1.100",
		TunnelPort:          2083,
		UUID:                "test-uuid-5678",
		FreeInterface:       "eth0",
		RestrictedInterface: "eth1",
	}

	data, err := GenerateClientConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateClientConfig failed: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated config is not valid JSON: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "eth0") {
		t.Error("config should contain free interface name")
	}
	if !strings.Contains(s, "eth1") {
		t.Error("config should contain restricted interface name")
	}
	if !strings.Contains(s, "192.168.1.100") {
		t.Error("config should contain server IP")
	}
	if !strings.Contains(s, "2083") {
		t.Error("config should contain tunnel port")
	}
	if !strings.Contains(s, "test-uuid-5678") {
		t.Error("config should contain UUID")
	}
	if !strings.Contains(s, "bridge") {
		t.Error("config should contain bridge tag")
	}
}
