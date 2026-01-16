//go:build linux

package platform

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LinuxNetworkManager implements NetworkManager for Linux
type LinuxNetworkManager struct {
	baseNetworkManager
}

// NewNetworkManager creates a new NetworkManager for the current platform
func NewNetworkManager() NetworkManager {
	return &LinuxNetworkManager{}
}

func (m *LinuxNetworkManager) ListInterfaces() ([]NetworkInterface, error) {
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
			IsEthernet:   !isWireless && strings.HasPrefix(iface.Name, "e"),
			MTU:          iface.MTU,
		}

		result = append(result, ni)
	}

	return result, nil
}

func (m *LinuxNetworkManager) GetInterface(name string) (*NetworkInterface, error) {
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
		IsEthernet:   !isWireless && strings.HasPrefix(iface.Name, "e"),
		MTU:          iface.MTU,
	}, nil
}

func (m *LinuxNetworkManager) ListUpInterfaces() ([]NetworkInterface, error) {
	all, err := m.ListInterfaces()
	if err != nil {
		return nil, err
	}
	return FilterUpInterfaces(all), nil
}

func (m *LinuxNetworkManager) PingTest(address string, iface string, timeout time.Duration) (time.Duration, error) {
	// Use ping command with interface binding
	args := []string{"-c", "1", "-W", fmt.Sprintf("%d", int(timeout.Seconds()))}
	if iface != "" {
		args = append(args, "-I", iface)
	}
	args = append(args, address)

	start := time.Now()
	cmd := exec.Command("ping", args...)
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		// Fallback to TCP ping
		port := "80" // Default HTTP port
		if strings.Contains(address, ":") {
			port = ""
		} else {
			address = address + ":" + port
		}
		return tcpPing(address, timeout)
	}

	return duration, nil
}

// isWireless checks if an interface is wireless by checking /sys/class/net/<iface>/wireless
func (m *LinuxNetworkManager) isWireless(name string) bool {
	wirelessPath := filepath.Join("/sys/class/net", name, "wireless")
	_, err := os.Stat(wirelessPath)
	return err == nil
}
