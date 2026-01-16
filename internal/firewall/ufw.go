package firewall

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// UFWManager manages UFW firewall rules
type UFWManager struct{}

// NewUFWManager creates a new UFWManager
func NewUFWManager() *UFWManager {
	return &UFWManager{}
}

func (m *UFWManager) Type() FirewallType {
	return FirewallTypeUFW
}

func (m *UFWManager) IsAvailable() bool {
	return isUFWAvailable()
}

func (m *UFWManager) IsEnabled() bool {
	output, err := exec.Command("ufw", "status").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "Status: active")
}

func (m *UFWManager) AllowPort(port uint16, protocol string, description string) error {
	// UFW syntax: ufw allow <port>/<protocol> comment '<description>'
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

	// Use --force to avoid interactive prompt
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

	// Look for the port/protocol in the output
	// UFW output format: [ 1] 8443/tcp                   ALLOW IN    Anywhere
	search := fmt.Sprintf("%d/%s", port, protocol)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, search) && strings.Contains(line, "ALLOW") {
			return true
		}
	}
	return false
}

// GetRuleNumber returns the rule number for a port, or 0 if not found
func (m *UFWManager) GetRuleNumber(port uint16, protocol string) int {
	output, err := exec.Command("ufw", "status", "numbered").Output()
	if err != nil {
		return 0
	}

	search := fmt.Sprintf("%d/%s", port, protocol)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, search) && strings.Contains(line, "ALLOW") {
			// Extract rule number from format "[ 1] ..."
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[") {
				end := strings.Index(line, "]")
				if end > 1 {
					numStr := strings.TrimSpace(line[1:end])
					if num, err := strconv.Atoi(numStr); err == nil {
						return num
					}
				}
			}
		}
	}
	return 0
}

// RemoveByRuleNumber removes a rule by its number
func (m *UFWManager) RemoveByRuleNumber(ruleNum int) error {
	cmd := exec.Command("ufw", "--force", "delete", strconv.Itoa(ruleNum))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ufw delete failed: %s", string(output))
	}
	return nil
}

// Status returns the UFW status output
func (m *UFWManager) Status() (string, error) {
	output, err := exec.Command("ufw", "status", "verbose").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get ufw status: %w", err)
	}
	return string(output), nil
}
