package util

import (
	"testing"
)

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		expected bool
	}{
		{"port 0 is invalid", 0, false},
		{"port 1 is valid", 1, true},
		{"port 80 is valid", 80, true},
		{"port 443 is valid", 443, true},
		{"port 1024 is valid", 1024, true},
		{"port 8080 is valid", 8080, true},
		{"port 49151 is valid", 49151, true},
		{"port 65535 is valid", 65535, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPort(tt.port)
			if result != tt.expected {
				t.Errorf("IsValidPort(%d) = %v, expected %v", tt.port, result, tt.expected)
			}
		})
	}
}

func TestIsRegisteredPort(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		expected bool
	}{
		{"port 0 is not registered", 0, false},
		{"port 80 is not registered", 80, false},
		{"port 443 is not registered", 443, false},
		{"port 1023 is not registered", 1023, false},
		{"port 1024 is registered", 1024, true},
		{"port 8080 is registered", 8080, true},
		{"port 49151 is registered", 49151, true},
		{"port 49152 is not registered", 49152, false},
		{"port 65535 is not registered", 65535, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRegisteredPort(tt.port)
			if result != tt.expected {
				t.Errorf("IsRegisteredPort(%d) = %v, expected %v", tt.port, result, tt.expected)
			}
		})
	}
}

func TestFindAvailablePort(t *testing.T) {
	// Test that FindAvailablePort returns a valid port
	port, err := FindAvailablePort(DefaultPortRetries)
	if err != nil {
		t.Fatalf("FindAvailablePort() returned error: %v", err)
	}

	if !IsRegisteredPort(port) {
		t.Errorf("FindAvailablePort() returned port %d which is not in registered range", port)
	}

	if !IsValidPort(port) {
		t.Errorf("FindAvailablePort() returned invalid port %d", port)
	}
}

func TestFindMultipleAvailablePorts(t *testing.T) {
	// Test finding multiple ports
	count := 3
	ports, err := FindMultipleAvailablePorts(count, count*DefaultPortRetries)
	if err != nil {
		t.Fatalf("FindMultipleAvailablePorts() returned error: %v", err)
	}

	if len(ports) != count {
		t.Errorf("FindMultipleAvailablePorts(%d) returned %d ports", count, len(ports))
	}

	// Check all ports are unique and valid
	seen := make(map[uint16]bool)
	for _, port := range ports {
		if !IsRegisteredPort(port) {
			t.Errorf("Port %d is not in registered range", port)
		}
		if seen[port] {
			t.Errorf("Duplicate port %d returned", port)
		}
		seen[port] = true
	}
}

func TestValidateOrSelectPort(t *testing.T) {
	// Test auto-selection when port is 0
	port, err := ValidateOrSelectPort(0, DefaultPortRetries)
	if err != nil {
		t.Fatalf("ValidateOrSelectPort(0) returned error: %v", err)
	}
	if !IsRegisteredPort(port) {
		t.Errorf("ValidateOrSelectPort(0) returned port %d which is not in registered range", port)
	}

	// Test with a likely available high port
	// Note: This test may be flaky if the port happens to be in use
	testPort := uint16(47123)
	result, err := ValidateOrSelectPort(testPort, DefaultPortRetries)
	if err == nil {
		if result != testPort {
			t.Errorf("ValidateOrSelectPort(%d) returned %d, expected same port", testPort, result)
		}
	}
	// If port is in use, the error is expected behavior
}

func TestPortRangeString(t *testing.T) {
	expected := "1-65535"
	result := PortRangeString()
	if result != expected {
		t.Errorf("PortRangeString() = %q, expected %q", result, expected)
	}
}

func TestRegisteredPortRangeString(t *testing.T) {
	expected := "1024-49151"
	result := RegisteredPortRangeString()
	if result != expected {
		t.Errorf("RegisteredPortRangeString() = %q, expected %q", result, expected)
	}
}

func TestPortConstants(t *testing.T) {
	// Verify constants are correct
	if MinRegisteredPort != 1024 {
		t.Errorf("MinRegisteredPort = %d, expected 1024", MinRegisteredPort)
	}
	if MaxRegisteredPort != 49151 {
		t.Errorf("MaxRegisteredPort = %d, expected 49151", MaxRegisteredPort)
	}
	if MinValidPort != 1 {
		t.Errorf("MinValidPort = %d, expected 1", MinValidPort)
	}
	if MaxValidPort != 65535 {
		t.Errorf("MaxValidPort = %d, expected 65535", MaxValidPort)
	}
	if DefaultPortRetries != 5 {
		t.Errorf("DefaultPortRetries = %d, expected 5", DefaultPortRetries)
	}
}
