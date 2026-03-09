package xray

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/net2share/nethopper/internal/config"
)

func TestGenerateServerConfig(t *testing.T) {
	cfg := &config.ServerConfig{
		SocksPort:  1080,
		TunnelPort: 2083,
		UUID:       "test-uuid-1234",
	}

	data, err := GenerateServerConfig(cfg)
	if err != nil {
		t.Fatalf("GenerateServerConfig failed: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated config is not valid JSON: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "1080") {
		t.Error("config should contain socks port")
	}
	if !strings.Contains(s, "2083") {
		t.Error("config should contain tunnel port")
	}
	if !strings.Contains(s, "test-uuid-1234") {
		t.Error("config should contain UUID")
	}
	if !strings.Contains(s, "portal") {
		t.Error("config should contain portal tag")
	}
	if !strings.Contains(s, "nethopper.internal") {
		t.Error("config should contain reverse domain")
	}
}
