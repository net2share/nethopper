//go:build darwin

package platform

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// DarwinNetworkManager implements NetworkManager for macOS
type DarwinNetworkManager struct {
	baseNetworkManager
}

// NewNetworkManager creates a new NetworkManager for the current platform
func NewNetworkManager() NetworkManager {
	return &DarwinNetworkManager{}
}

func (m *DarwinNetworkManager) ListInterfaces() ([]NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}

	result := make([]NetworkInterface, 0, len(ifaces))
	for _, iface := range ifaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		ipv4, ipv6 := getInterfaceAddrs(iface)
		isWireless := m.isWireless(iface.Name)

		ni := NetworkInterface{
			Name:         iface.Name,
			Index:        iface.Index,
			HardwareAddr: iface.HardwareAddr.String(),
			IPv4Addrs:    ipv4,
			IPv6Addrs:    ipv6,
			IsUp:         iface.Flags&net.FlagUp != 0,
			IsLoopback:   false,
			IsWireless:   isWireless,
			IsEthernet:   !isWireless && strings.HasPrefix(iface.Name, "en"),
			MTU:          iface.MTU,
		}

		result = append(result, ni)
	}

	return result, nil
}

func (m *DarwinNetworkManager) GetInterface(name string) (*NetworkInterface, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("interface '%s' not found: %w", name, err)
	}

	ipv4, ipv6 := getInterfaceAddrs(*iface)
	isWireless := m.isWireless(iface.Name)

	return &NetworkInterface{
		Name:         iface.Name,
		Index:        iface.Index,
		HardwareAddr: iface.HardwareAddr.String(),
		IPv4Addrs:    ipv4,
		IPv6Addrs:    ipv6,
		IsUp:         iface.Flags&net.FlagUp != 0,
		IsLoopback:   iface.Flags&net.FlagLoopback != 0,
		IsWireless:   isWireless,
		IsEthernet:   !isWireless && strings.HasPrefix(iface.Name, "en"),
		MTU:          iface.MTU,
	}, nil
}

func (m *DarwinNetworkManager) ListUpInterfaces() ([]NetworkInterface, error) {
	all, err := m.ListInterfaces()
	if err != nil {
		return nil, err
	}
	return FilterUpInterfaces(all), nil
}

func (m *DarwinNetworkManager) PingTest(address string, iface string, timeout time.Duration) (time.Duration, error) {
	// Use ping command with interface binding
	args := []string{"-c", "1", "-W", fmt.Sprintf("%d", int(timeout.Milliseconds()))}
	if iface != "" {
		args = append(args, "-b", iface)
	}
	args = append(args, address)

	start := time.Now()
	cmd := exec.Command("ping", args...)
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		// Fallback to TCP ping
		port := "80"
		if !strings.Contains(address, ":") {
			address = address + ":" + port
		}
		return tcpPing(address, timeout)
	}

	return duration, nil
}

// isWireless checks if an interface is wireless
// On macOS, Wi-Fi interfaces typically have names starting with "en" and are detected via airport
func (m *DarwinNetworkManager) isWireless(name string) bool {
	// Common Wi-Fi interface pattern on macOS
	// en0 is typically the first Wi-Fi interface
	// We can also try to detect using networksetup
	if !strings.HasPrefix(name, "en") {
		return false
	}

	// Try to detect using networksetup
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: assume en0 is Wi-Fi (common on Macs)
		return name == "en0"
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if strings.Contains(line, "Wi-Fi") || strings.Contains(line, "Airport") {
			// Check next line for device name
			if i+1 < len(lines) && strings.Contains(lines[i+1], name) {
				return true
			}
		}
	}

	return false
}
