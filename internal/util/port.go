package util

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	// Port ranges
	MinRegisteredPort = 1024
	MaxRegisteredPort = 49151
	MinValidPort      = 1
	MaxValidPort      = 65535

	// Default retry count
	DefaultPortRetries = 5
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// IsPortAvailable checks if a TCP port is available on all interfaces
func IsPortAvailable(port uint16) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// IsPortAvailableOnInterface checks if a TCP port is available on a specific interface
func IsPortAvailableOnInterface(port uint16, iface string) bool {
	addr := fmt.Sprintf("%s:%d", iface, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// IsValidPort checks if a port number is valid
func IsValidPort(port uint16) bool {
	return port >= MinValidPort && port <= MaxValidPort
}

// IsRegisteredPort checks if a port is in the registered range (1024-49151)
func IsRegisteredPort(port uint16) bool {
	return port >= MinRegisteredPort && port <= MaxRegisteredPort
}

// FindAvailablePort finds an available port in the registered range
// Returns 0 if no port is found after maxRetries attempts
func FindAvailablePort(maxRetries int) (uint16, error) {
	if maxRetries <= 0 {
		maxRetries = DefaultPortRetries
	}

	for i := 0; i < maxRetries; i++ {
		port := uint16(rng.Intn(MaxRegisteredPort-MinRegisteredPort+1) + MinRegisteredPort)
		if IsPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("failed to find available port after %d attempts", maxRetries)
}

// FindMultipleAvailablePorts finds multiple available ports
// Returns a slice of available ports, or an error if not enough ports can be found
func FindMultipleAvailablePorts(count int, maxRetries int) ([]uint16, error) {
	if maxRetries <= 0 {
		maxRetries = DefaultPortRetries * count
	}

	ports := make([]uint16, 0, count)
	usedPorts := make(map[uint16]bool)

	for attempt := 0; attempt < maxRetries && len(ports) < count; attempt++ {
		port := uint16(rng.Intn(MaxRegisteredPort-MinRegisteredPort+1) + MinRegisteredPort)
		if usedPorts[port] {
			continue
		}
		if IsPortAvailable(port) {
			ports = append(ports, port)
			usedPorts[port] = true
		}
	}

	if len(ports) < count {
		return ports, fmt.Errorf("could only find %d available ports out of %d requested", len(ports), count)
	}

	return ports, nil
}

// ValidateOrSelectPort validates a user-provided port or selects a random one
// If port is 0, finds an available port
// If port is non-zero, validates it's available
func ValidateOrSelectPort(port uint16, maxRetries int) (uint16, error) {
	if port == 0 {
		return FindAvailablePort(maxRetries)
	}

	if !IsValidPort(port) {
		return 0, fmt.Errorf("port %d is out of valid range (%d-%d)", port, MinValidPort, MaxValidPort)
	}

	if !IsPortAvailable(port) {
		return 0, fmt.Errorf("port %d is not available", port)
	}

	return port, nil
}

// PortRangeString returns a human-readable string for the valid port range
func PortRangeString() string {
	return fmt.Sprintf("%d-%d", MinValidPort, MaxValidPort)
}

// RegisteredPortRangeString returns a human-readable string for the registered port range
func RegisteredPortRangeString() string {
	return fmt.Sprintf("%d-%d", MinRegisteredPort, MaxRegisteredPort)
}
