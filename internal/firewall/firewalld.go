package firewall

import (
	"fmt"
	"os/exec"
	"strings"
)

// FirewalldManager manages firewalld firewall rules
type FirewalldManager struct{}

// NewFirewalldManager creates a new FirewalldManager
func NewFirewalldManager() *FirewalldManager {
	return &FirewalldManager{}
}

func (m *FirewalldManager) Type() FirewallType {
	return FirewallTypeFirewalld
}

func (m *FirewalldManager) IsAvailable() bool {
	return isFirewalldAvailable()
}

func (m *FirewalldManager) IsEnabled() bool {
	output, err := exec.Command("firewall-cmd", "--state").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "running"
}

func (m *FirewalldManager) AllowPort(port uint16, protocol string, description string) error {
	// firewall-cmd --permanent --add-port=<port>/<protocol>
	portStr := fmt.Sprintf("%d/%s", port, protocol)

	// Add permanent rule
	cmd := exec.Command("firewall-cmd", "--permanent", "--add-port="+portStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd add-port failed: %s", string(output))
	}

	// Reload to apply
	if err := m.Reload(); err != nil {
		return fmt.Errorf("failed to reload firewalld: %w", err)
	}

	return nil
}

func (m *FirewalldManager) RemovePort(port uint16, protocol string) error {
	portStr := fmt.Sprintf("%d/%s", port, protocol)

	// Remove permanent rule
	cmd := exec.Command("firewall-cmd", "--permanent", "--remove-port="+portStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd remove-port failed: %s", string(output))
	}

	// Reload to apply
	if err := m.Reload(); err != nil {
		return fmt.Errorf("failed to reload firewalld: %w", err)
	}

	return nil
}

func (m *FirewalldManager) IsPortAllowed(port uint16, protocol string) bool {
	portStr := fmt.Sprintf("%d/%s", port, protocol)

	// Check both runtime and permanent
	cmd := exec.Command("firewall-cmd", "--query-port="+portStr)
	err := cmd.Run()
	if err == nil {
		return true
	}

	// Also check permanent rules
	cmd = exec.Command("firewall-cmd", "--permanent", "--query-port="+portStr)
	err = cmd.Run()
	return err == nil
}

// Reload reloads firewalld configuration
func (m *FirewalldManager) Reload() error {
	cmd := exec.Command("firewall-cmd", "--reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd reload failed: %s", string(output))
	}
	return nil
}

// ListPorts returns a list of allowed ports
func (m *FirewalldManager) ListPorts() ([]string, error) {
	output, err := exec.Command("firewall-cmd", "--list-ports").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %w", err)
	}

	portStr := strings.TrimSpace(string(output))
	if portStr == "" {
		return []string{}, nil
	}

	return strings.Split(portStr, " "), nil
}

// GetZone returns the current default zone
func (m *FirewalldManager) GetZone() (string, error) {
	output, err := exec.Command("firewall-cmd", "--get-default-zone").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get zone: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// AllowPortInZone allows a port in a specific zone
func (m *FirewalldManager) AllowPortInZone(zone string, port uint16, protocol string) error {
	portStr := fmt.Sprintf("%d/%s", port, protocol)

	cmd := exec.Command("firewall-cmd", "--permanent", "--zone="+zone, "--add-port="+portStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd add-port failed: %s", string(output))
	}

	return m.Reload()
}

// Status returns the firewalld status output
func (m *FirewalldManager) Status() (string, error) {
	output, err := exec.Command("firewall-cmd", "--state").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get firewalld state: %w", err)
	}

	state := strings.TrimSpace(string(output))

	// Get more details
	zoneOutput, _ := exec.Command("firewall-cmd", "--get-default-zone").Output()
	zone := strings.TrimSpace(string(zoneOutput))

	portsOutput, _ := exec.Command("firewall-cmd", "--list-ports").Output()
	ports := strings.TrimSpace(string(portsOutput))

	return fmt.Sprintf("State: %s\nDefault Zone: %s\nOpen Ports: %s", state, zone, ports), nil
}
