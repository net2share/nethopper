package firewall

import (
	"fmt"
	"os/exec"
	"strings"
)

type FirewalldManager struct{}

func (m *FirewalldManager) Type() FirewallType { return FirewallTypeFirewalld }

func (m *FirewalldManager) IsEnabled() bool {
	output, err := exec.Command("firewall-cmd", "--state").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "running"
}

func (m *FirewalldManager) AllowPort(port uint16, protocol string, description string) error {
	portStr := fmt.Sprintf("%d/%s", port, protocol)
	cmd := exec.Command("firewall-cmd", "--permanent", "--add-port="+portStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd add-port failed: %s", string(output))
	}
	return m.reload()
}

func (m *FirewalldManager) RemovePort(port uint16, protocol string) error {
	portStr := fmt.Sprintf("%d/%s", port, protocol)
	cmd := exec.Command("firewall-cmd", "--permanent", "--remove-port="+portStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd remove-port failed: %s", string(output))
	}
	return m.reload()
}

func (m *FirewalldManager) IsPortAllowed(port uint16, protocol string) bool {
	portStr := fmt.Sprintf("%d/%s", port, protocol)
	if exec.Command("firewall-cmd", "--query-port="+portStr).Run() == nil {
		return true
	}
	return exec.Command("firewall-cmd", "--permanent", "--query-port="+portStr).Run() == nil
}

func (m *FirewalldManager) reload() error {
	cmd := exec.Command("firewall-cmd", "--reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("firewall-cmd reload failed: %s", string(output))
	}
	return nil
}
