package firewall

import (
	"fmt"
	"os/exec"
)

type FirewallType string

const (
	FirewallTypeNone      FirewallType = "none"
	FirewallTypeUFW       FirewallType = "ufw"
	FirewallTypeFirewalld FirewallType = "firewalld"
)

type FirewallManager interface {
	Type() FirewallType
	IsEnabled() bool
	AllowPort(port uint16, protocol string, description string) error
	RemovePort(port uint16, protocol string) error
	IsPortAllowed(port uint16, protocol string) bool
}

// DetectFirewall detects the available firewall on the system.
func DetectFirewall() FirewallManager {
	if _, err := exec.LookPath("ufw"); err == nil {
		return &UFWManager{}
	}
	if _, err := exec.LookPath("firewall-cmd"); err == nil {
		return &FirewalldManager{}
	}
	return &NoopManager{}
}

// AllowPorts allows multiple ports through the firewall.
func AllowPorts(fm FirewallManager, ports []uint16, protocol, description string) error {
	if fm.Type() == FirewallTypeNone || !fm.IsEnabled() {
		return nil
	}
	for _, port := range ports {
		if fm.IsPortAllowed(port, protocol) {
			continue
		}
		desc := fmt.Sprintf("%s (port %d)", description, port)
		if err := fm.AllowPort(port, protocol, desc); err != nil {
			return fmt.Errorf("failed to allow port %d: %w", port, err)
		}
	}
	return nil
}

// RemovePorts removes multiple port rules from the firewall.
func RemovePorts(fm FirewallManager, ports []uint16, protocol string) error {
	if fm.Type() == FirewallTypeNone {
		return nil
	}
	for _, port := range ports {
		if !fm.IsPortAllowed(port, protocol) {
			continue
		}
		if err := fm.RemovePort(port, protocol); err != nil {
			return fmt.Errorf("failed to remove port %d: %w", port, err)
		}
	}
	return nil
}

// NoopManager is a no-op firewall manager.
type NoopManager struct{}

func (m *NoopManager) Type() FirewallType                                        { return FirewallTypeNone }
func (m *NoopManager) IsEnabled() bool                                           { return false }
func (m *NoopManager) AllowPort(port uint16, protocol string, desc string) error { return nil }
func (m *NoopManager) RemovePort(port uint16, protocol string) error             { return nil }
func (m *NoopManager) IsPortAllowed(port uint16, protocol string) bool           { return true }
