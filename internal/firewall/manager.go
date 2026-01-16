package firewall

import (
	"fmt"
	"os/exec"
)

// FirewallType represents the type of firewall
type FirewallType string

const (
	FirewallTypeNone     FirewallType = "none"
	FirewallTypeUFW      FirewallType = "ufw"
	FirewallTypeFirewalld FirewallType = "firewalld"
)

// FirewallManager defines the interface for firewall operations
type FirewallManager interface {
	// Type returns the type of firewall
	Type() FirewallType

	// IsAvailable checks if this firewall is available on the system
	IsAvailable() bool

	// IsEnabled checks if the firewall is enabled
	IsEnabled() bool

	// AllowPort allows a port through the firewall
	AllowPort(port uint16, protocol string, description string) error

	// RemovePort removes a port rule from the firewall
	RemovePort(port uint16, protocol string) error

	// IsPortAllowed checks if a port is already allowed
	IsPortAllowed(port uint16, protocol string) bool
}

// DetectFirewall detects the available firewall on the system
func DetectFirewall() FirewallManager {
	// Check for UFW first (Debian/Ubuntu)
	if isUFWAvailable() {
		return NewUFWManager()
	}

	// Check for firewalld (RHEL/CentOS/Fedora)
	if isFirewalldAvailable() {
		return NewFirewalldManager()
	}

	// No firewall detected
	return NewNoopManager()
}

// isUFWAvailable checks if UFW is available
func isUFWAvailable() bool {
	_, err := exec.LookPath("ufw")
	return err == nil
}

// isFirewalldAvailable checks if firewalld is available
func isFirewalldAvailable() bool {
	_, err := exec.LookPath("firewall-cmd")
	return err == nil
}

// AllowPorts is a helper function to allow multiple ports
func AllowPorts(fm FirewallManager, ports []uint16, protocol, description string) error {
	if fm.Type() == FirewallTypeNone || !fm.IsEnabled() {
		return nil
	}

	for _, port := range ports {
		if fm.IsPortAllowed(port, protocol) {
			continue // Already allowed
		}
		desc := fmt.Sprintf("%s (port %d)", description, port)
		if err := fm.AllowPort(port, protocol, desc); err != nil {
			return fmt.Errorf("failed to allow port %d: %w", port, err)
		}
	}
	return nil
}

// RemovePorts is a helper function to remove multiple port rules
func RemovePorts(fm FirewallManager, ports []uint16, protocol string) error {
	if fm.Type() == FirewallTypeNone {
		return nil
	}

	for _, port := range ports {
		if !fm.IsPortAllowed(port, protocol) {
			continue // Not allowed, nothing to remove
		}
		if err := fm.RemovePort(port, protocol); err != nil {
			return fmt.Errorf("failed to remove port %d: %w", port, err)
		}
	}
	return nil
}

// NoopManager is a no-op firewall manager for systems without a firewall
type NoopManager struct{}

// NewNoopManager creates a new NoopManager
func NewNoopManager() *NoopManager {
	return &NoopManager{}
}

func (m *NoopManager) Type() FirewallType {
	return FirewallTypeNone
}

func (m *NoopManager) IsAvailable() bool {
	return false
}

func (m *NoopManager) IsEnabled() bool {
	return false
}

func (m *NoopManager) AllowPort(port uint16, protocol string, description string) error {
	return nil
}

func (m *NoopManager) RemovePort(port uint16, protocol string) error {
	return nil
}

func (m *NoopManager) IsPortAllowed(port uint16, protocol string) bool {
	return true // No firewall means all ports are "allowed"
}
