//go:build windows

package platform

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// WindowsNetworkManager implements NetworkManager for Windows
type WindowsNetworkManager struct {
	baseNetworkManager
}

// NewNetworkManager creates a new NetworkManager for the current platform
func NewNetworkManager() NetworkManager {
	return &WindowsNetworkManager{}
}

func (m *WindowsNetworkManager) ListInterfaces() ([]NetworkInterface, error) {
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
			IsEthernet:   !isWireless,
			MTU:          iface.MTU,
		}

		result = append(result, ni)
	}

	return result, nil
}

func (m *WindowsNetworkManager) GetInterface(name string) (*NetworkInterface, error) {
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
		IsEthernet:   !isWireless,
		MTU:          iface.MTU,
	}, nil
}

func (m *WindowsNetworkManager) ListUpInterfaces() ([]NetworkInterface, error) {
	all, err := m.ListInterfaces()
	if err != nil {
		return nil, err
	}
	return FilterUpInterfaces(all), nil
}

func (m *WindowsNetworkManager) PingTest(address string, iface string, timeout time.Duration) (time.Duration, error) {
	// Use ping command
	// Windows ping uses -n for count and -w for timeout in milliseconds
	args := []string{"-n", "1", "-w", fmt.Sprintf("%d", int(timeout.Milliseconds()))}

	// Note: Windows ping doesn't have a simple interface binding option
	// The source address needs to be specified with -S
	// For now, we'll use the default interface
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

// isWireless checks if an interface is wireless on Windows
func (m *WindowsNetworkManager) isWireless(name string) bool {
	// Use netsh to check interface type
	// netsh wlan show interfaces will list wireless interfaces
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.Output()
	if err != nil {
		// If netsh wlan fails, assume not wireless
		return false
	}

	// Check if interface name appears in wireless interfaces output
	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(name))
}

// GetInterfaceByGUID returns an interface by its GUID (Windows-specific)
func (m *WindowsNetworkManager) GetInterfaceByGUID(guid string) (*NetworkInterface, error) {
	ifaces, err := m.ListInterfaces()
	if err != nil {
		return nil, err
	}

	// On Windows, the interface name is often the GUID
	for _, iface := range ifaces {
		if iface.Name == guid {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("interface with GUID '%s' not found", guid)
}
