package config

import (
	"encoding/json"
	"testing"
)

func TestGenerateServerConfig(t *testing.T) {
	config := &ServerConfig{
		ListenPort:             2083,
		MTProtoPort:            8443,
		VMESSPort:              8081,
		MTProtoSecret:          "0123456789abcdef0123456789abcdef",
		VMESSUuid:              "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FallbackHost:           "storage.googleapis.com",
		MultiplexPerConnection: 50,
		LogLevel:               "warn",
	}

	result, err := GenerateServerConfig(config)
	if err != nil {
		t.Fatalf("GenerateServerConfig() returned error: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Generated config is not valid JSON: %v", err)
	}

	// Check log level
	log, ok := parsed["log"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'log' section")
	}
	if log["level"] != "warn" {
		t.Errorf("Log level = %v, expected %q", log["level"], "warn")
	}

	// Check endpoints
	endpoints, ok := parsed["endpoints"].([]interface{})
	if !ok || len(endpoints) < 1 {
		t.Fatal("Missing 'endpoints' section")
	}
	ep := endpoints[0].(map[string]interface{})
	if ep["type"] != "reverse" {
		t.Errorf("Endpoint type = %v, expected %q", ep["type"], "reverse")
	}
	if ep["listen_port"].(float64) != 2083 {
		t.Errorf("Endpoint listen_port = %v, expected 2083", ep["listen_port"])
	}

	// Check inbounds
	inbounds, ok := parsed["inbounds"].([]interface{})
	if !ok || len(inbounds) < 1 {
		t.Fatal("Missing 'inbounds' section")
	}
	inbound := inbounds[0].(map[string]interface{})
	if inbound["type"] != "mtproto" {
		t.Errorf("Inbound type = %v, expected %q", inbound["type"], "mtproto")
	}
	if inbound["listen_port"].(float64) != 8443 {
		t.Errorf("Inbound listen_port = %v, expected 8443", inbound["listen_port"])
	}
	if inbound["fallback_host"] != "storage.googleapis.com" {
		t.Errorf("Inbound fallback_host = %v, expected %q", inbound["fallback_host"], "storage.googleapis.com")
	}

	// Check users in inbound
	users := inbound["users"].([]interface{})
	if len(users) < 1 {
		t.Fatal("Missing users in inbound")
	}
	user := users[0].(map[string]interface{})
	if user["secret"] != config.MTProtoSecret {
		t.Errorf("User secret = %v, expected %q", user["secret"], config.MTProtoSecret)
	}

	// Check outbounds
	outbounds, ok := parsed["outbounds"].([]interface{})
	if !ok || len(outbounds) < 2 {
		t.Fatal("Missing 'outbounds' section or insufficient outbounds")
	}
	vmessOut := outbounds[0].(map[string]interface{})
	if vmessOut["type"] != "vmess" {
		t.Errorf("Outbound type = %v, expected %q", vmessOut["type"], "vmess")
	}
	if vmessOut["uuid"] != config.VMESSUuid {
		t.Errorf("Outbound uuid = %v, expected %q", vmessOut["uuid"], config.VMESSUuid)
	}
	if vmessOut["server_port"].(float64) != 8081 {
		t.Errorf("Outbound server_port = %v, expected 8081", vmessOut["server_port"])
	}
}

func TestGenerateServerConfigValidation(t *testing.T) {
	// Test with invalid config (missing required fields)
	invalidConfig := &ServerConfig{
		ListenPort: 0, // Missing required field
	}

	_, err := GenerateServerConfig(invalidConfig)
	if err == nil {
		t.Error("GenerateServerConfig() should have returned error for invalid config")
	}
}

func TestGenerateFreenetConfig(t *testing.T) {
	config := &FreenetConfig{
		ServerIP:            "192.168.1.100",
		ServerPort:          2083,
		VMESSPort:           8081,
		VMESSUuid:           "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FreeInterface:       "eth0",
		RestrictedInterface: "wlan0",
		MaxConnections:      30,
		LogLevel:            "debug",
	}

	result, err := GenerateFreenetConfig(config)
	if err != nil {
		t.Fatalf("GenerateFreenetConfig() returned error: %v", err)
	}

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Generated config is not valid JSON: %v", err)
	}

	// Check log level
	log, ok := parsed["log"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'log' section")
	}
	if log["level"] != "debug" {
		t.Errorf("Log level = %v, expected %q", log["level"], "debug")
	}

	// Check endpoints
	endpoints, ok := parsed["endpoints"].([]interface{})
	if !ok || len(endpoints) < 1 {
		t.Fatal("Missing 'endpoints' section")
	}
	ep := endpoints[0].(map[string]interface{})
	if ep["type"] != "reverse" {
		t.Errorf("Endpoint type = %v, expected %q", ep["type"], "reverse")
	}
	if ep["server"] != "192.168.1.100" {
		t.Errorf("Endpoint server = %v, expected %q", ep["server"], "192.168.1.100")
	}
	if ep["server_port"].(float64) != 2083 {
		t.Errorf("Endpoint server_port = %v, expected 2083", ep["server_port"])
	}

	// Check multiplex config
	multiplex := ep["multiplex"].(map[string]interface{})
	if multiplex["max_connections"].(float64) != 30 {
		t.Errorf("Multiplex max_connections = %v, expected 30", multiplex["max_connections"])
	}

	// Check inbounds
	inbounds, ok := parsed["inbounds"].([]interface{})
	if !ok || len(inbounds) < 1 {
		t.Fatal("Missing 'inbounds' section")
	}
	inbound := inbounds[0].(map[string]interface{})
	if inbound["type"] != "vmess" {
		t.Errorf("Inbound type = %v, expected %q", inbound["type"], "vmess")
	}
	if inbound["listen_port"].(float64) != 8081 {
		t.Errorf("Inbound listen_port = %v, expected 8081", inbound["listen_port"])
	}

	// Check users in inbound
	users := inbound["users"].([]interface{})
	if len(users) < 1 {
		t.Fatal("Missing users in inbound")
	}
	user := users[0].(map[string]interface{})
	if user["uuid"] != config.VMESSUuid {
		t.Errorf("User uuid = %v, expected %q", user["uuid"], config.VMESSUuid)
	}

	// Check outbounds
	outbounds, ok := parsed["outbounds"].([]interface{})
	if !ok || len(outbounds) < 2 {
		t.Fatal("Missing 'outbounds' section or insufficient outbounds")
	}

	// Free interface outbound
	freeOut := outbounds[0].(map[string]interface{})
	if freeOut["type"] != "direct" {
		t.Errorf("Free outbound type = %v, expected %q", freeOut["type"], "direct")
	}
	if freeOut["bind_interface"] != "eth0" {
		t.Errorf("Free outbound bind_interface = %v, expected %q", freeOut["bind_interface"], "eth0")
	}
	if freeOut["tag"] != "free" {
		t.Errorf("Free outbound tag = %v, expected %q", freeOut["tag"], "free")
	}

	// Restricted interface outbound
	restrictedOut := outbounds[1].(map[string]interface{})
	if restrictedOut["type"] != "direct" {
		t.Errorf("Restricted outbound type = %v, expected %q", restrictedOut["type"], "direct")
	}
	if restrictedOut["bind_interface"] != "wlan0" {
		t.Errorf("Restricted outbound bind_interface = %v, expected %q", restrictedOut["bind_interface"], "wlan0")
	}
	if restrictedOut["tag"] != "filternet" {
		t.Errorf("Restricted outbound tag = %v, expected %q", restrictedOut["tag"], "filternet")
	}
}

func TestGenerateFreenetConfigValidation(t *testing.T) {
	// Test with invalid config (missing required fields)
	invalidConfig := &FreenetConfig{
		ServerIP: "", // Missing required field
	}

	_, err := GenerateFreenetConfig(invalidConfig)
	if err == nil {
		t.Error("GenerateFreenetConfig() should have returned error for invalid config")
	}
}

func TestGeneratedConfigIsValidJSON(t *testing.T) {
	// Test that generated configs are always valid JSON
	serverConfig := &ServerConfig{
		ListenPort:             2083,
		MTProtoPort:            8443,
		VMESSPort:              8081,
		MTProtoSecret:          "0123456789abcdef0123456789abcdef",
		VMESSUuid:              "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FallbackHost:           "example.com",
		MultiplexPerConnection: 100,
		LogLevel:               "trace",
	}

	serverResult, err := GenerateServerConfig(serverConfig)
	if err != nil {
		t.Fatalf("GenerateServerConfig() returned error: %v", err)
	}

	if !json.Valid(serverResult) {
		t.Error("Server config is not valid JSON")
	}

	freenetConfig := &FreenetConfig{
		ServerIP:            "10.0.0.1",
		ServerPort:          1234,
		VMESSPort:           5678,
		VMESSUuid:           "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		FreeInterface:       "enp0s1",
		RestrictedInterface: "wlan1",
		MaxConnections:      50,
		LogLevel:            "error",
	}

	freenetResult, err := GenerateFreenetConfig(freenetConfig)
	if err != nil {
		t.Fatalf("GenerateFreenetConfig() returned error: %v", err)
	}

	if !json.Valid(freenetResult) {
		t.Error("Freenet config is not valid JSON")
	}
}

func TestGeneratedConfigPreservesPrettyFormat(t *testing.T) {
	config := &ServerConfig{
		ListenPort:             2083,
		MTProtoPort:            8443,
		VMESSPort:              8081,
		MTProtoSecret:          "0123456789abcdef0123456789abcdef",
		VMESSUuid:              "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FallbackHost:           "storage.googleapis.com",
		MultiplexPerConnection: 50,
		LogLevel:               "warn",
	}

	result, err := GenerateServerConfig(config)
	if err != nil {
		t.Fatalf("GenerateServerConfig() returned error: %v", err)
	}

	// Check that it's indented (pretty printed)
	resultStr := string(result)
	if resultStr[0] != '{' {
		t.Error("Generated config should start with '{'")
	}

	// Should contain newlines (not minified)
	if len(resultStr) > 0 && !containsNewline(resultStr) {
		t.Error("Generated config should be pretty-printed with newlines")
	}
}

func containsNewline(s string) bool {
	for _, c := range s {
		if c == '\n' {
			return true
		}
	}
	return false
}

func TestGenerateConfigWithDisabledLogging(t *testing.T) {
	// Test server config with disabled logging
	serverConfig := &ServerConfig{
		ListenPort:             2083,
		MTProtoPort:            8443,
		VMESSPort:              8081,
		MTProtoSecret:          "0123456789abcdef0123456789abcdef",
		VMESSUuid:              "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FallbackHost:           "storage.googleapis.com",
		MultiplexPerConnection: 50,
		LogLevel:               "disabled",
	}

	serverResult, err := GenerateServerConfig(serverConfig)
	if err != nil {
		t.Fatalf("GenerateServerConfig() with disabled logging returned error: %v", err)
	}

	var serverParsed map[string]interface{}
	if err := json.Unmarshal(serverResult, &serverParsed); err != nil {
		t.Fatalf("Generated server config is not valid JSON: %v", err)
	}

	serverLog, ok := serverParsed["log"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'log' section in server config")
	}
	if disabled, ok := serverLog["disabled"].(bool); !ok || !disabled {
		t.Errorf("Server log.disabled = %v, expected true", serverLog["disabled"])
	}
	if _, hasLevel := serverLog["level"]; hasLevel {
		t.Error("Server config should not have 'level' when logging is disabled")
	}

	// Test freenet config with disabled logging
	freenetConfig := &FreenetConfig{
		ServerIP:            "192.168.1.100",
		ServerPort:          2083,
		VMESSPort:           8081,
		VMESSUuid:           "103b0aae-3384-4d23-9f5b-2d15be377a23",
		FreeInterface:       "eth0",
		RestrictedInterface: "wlan0",
		MaxConnections:      30,
		LogLevel:            "disabled",
	}

	freenetResult, err := GenerateFreenetConfig(freenetConfig)
	if err != nil {
		t.Fatalf("GenerateFreenetConfig() with disabled logging returned error: %v", err)
	}

	var freenetParsed map[string]interface{}
	if err := json.Unmarshal(freenetResult, &freenetParsed); err != nil {
		t.Fatalf("Generated freenet config is not valid JSON: %v", err)
	}

	freenetLog, ok := freenetParsed["log"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'log' section in freenet config")
	}
	if disabled, ok := freenetLog["disabled"].(bool); !ok || !disabled {
		t.Errorf("Freenet log.disabled = %v, expected true", freenetLog["disabled"])
	}
	if _, hasLevel := freenetLog["level"]; hasLevel {
		t.Error("Freenet config should not have 'level' when logging is disabled")
	}
}
