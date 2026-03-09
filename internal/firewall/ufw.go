package firewall

import (
	"fmt"
	"os/exec"
	"strings"
)

type UFWManager struct{}

func (m *UFWManager) Type() FirewallType { return FirewallTypeUFW }

func (m *UFWManager) IsEnabled() bool {
	output, err := exec.Command("ufw", "status").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Status: active")
}

func (m *UFWManager) AllowPort(port uint16, protocol string, description string) error {
	rule := fmt.Sprintf("%d/%s", port, protocol)
	var cmd *exec.Cmd
	if description != "" {
		cmd = exec.Command("ufw", "allow", rule, "comment", description)
	} else {
		cmd = exec.Command("ufw", "allow", rule)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ufw allow failed: %s", string(output))
	}
	return nil
}

func (m *UFWManager) RemovePort(port uint16, protocol string) error {
	rule := fmt.Sprintf("%d/%s", port, protocol)
	cmd := exec.Command("ufw", "--force", "delete", "allow", rule)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ufw delete failed: %s", string(output))
	}
	return nil
}

func (m *UFWManager) IsPortAllowed(port uint16, protocol string) bool {
	output, err := exec.Command("ufw", "status", "numbered").Output()
	if err != nil {
		return false
	}
	search := fmt.Sprintf("%d/%s", port, protocol)
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, search) && strings.Contains(line, "ALLOW") {
			return true
		}
	}
	return false
}
