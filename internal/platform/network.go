package platform

import (
	"fmt"
	"net"
	"time"
)

// NetworkInterface represents a network interface with its properties
type NetworkInterface struct {
	Name         string   // Interface name (e.g., eth0, en0)
	Index        int      // System index
	HardwareAddr string   // MAC address
	IPv4Addrs    []string // IPv4 addresses
	IPv6Addrs    []string // IPv6 addresses
	IsUp         bool     // Interface is up
	IsLoopback   bool     // Is loopback interface
	IsWireless   bool     // Is wireless interface (platform-specific detection)
	IsEthernet   bool     // Is ethernet interface
	MTU          int      // Maximum transmission unit
}

// NetworkManager provides network interface operations
type NetworkManager interface {
	// ListInterfaces returns all non-loopback network interfaces
	ListInterfaces() ([]NetworkInterface, error)

	// GetInterface returns a specific interface by name
	GetInterface(name string) (*NetworkInterface, error)

	// ListUpInterfaces returns only interfaces that are up
	ListUpInterfaces() ([]NetworkInterface, error)

	// PingTest performs a ping test to the given address
	PingTest(address string, iface string, timeout time.Duration) (time.Duration, error)
}

// String returns a human-readable representation of the interface
func (ni *NetworkInterface) String() string {
	status := "DOWN"
	if ni.IsUp {
		status = "UP"
	}

	ifaceType := "Unknown"
	if ni.IsWireless {
		ifaceType = "Wireless"
	} else if ni.IsEthernet {
		ifaceType = "Ethernet"
	}

	addrs := ""
	if len(ni.IPv4Addrs) > 0 {
		addrs = ni.IPv4Addrs[0]
	}

	return fmt.Sprintf("%s (%s) - %s [%s]", ni.Name, addrs, ifaceType, status)
}

// HasIPv4 returns true if the interface has at least one IPv4 address
func (ni *NetworkInterface) HasIPv4() bool {
	return len(ni.IPv4Addrs) > 0
}

// HasIPv6 returns true if the interface has at least one IPv6 address
func (ni *NetworkInterface) HasIPv6() bool {
	return len(ni.IPv6Addrs) > 0
}

// baseNetworkManager provides common functionality for all platforms
type baseNetworkManager struct{}

// getInterfaceAddrs extracts IPv4 and IPv6 addresses from an interface
func getInterfaceAddrs(iface net.Interface) ([]string, []string) {
	var ipv4, ipv6 []string

	addrs, err := iface.Addrs()
	if err != nil {
		return ipv4, ipv6
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil {
			continue
		}

		if ip4 := ip.To4(); ip4 != nil {
			ipv4 = append(ipv4, ip4.String())
		} else if ip6 := ip.To16(); ip6 != nil {
			// Skip link-local addresses for display
			if !ip6.IsLinkLocalUnicast() {
				ipv6 = append(ipv6, ip6.String())
			}
		}
	}

	return ipv4, ipv6
}

// tcpPing performs a TCP ping to the given address
func tcpPing(address string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start), nil
}

// ValidateInterfaceSelection validates that the selected interfaces are valid
func ValidateInterfaceSelection(interfaces []NetworkInterface, freeIface, restrictedIface string) error {
	if freeIface == restrictedIface {
		return fmt.Errorf("free and restricted interfaces must be different")
	}

	var foundFree, foundRestricted bool
	for _, iface := range interfaces {
		if iface.Name == freeIface {
			foundFree = true
			if !iface.IsUp {
				return fmt.Errorf("free interface '%s' is down", freeIface)
			}
		}
		if iface.Name == restrictedIface {
			foundRestricted = true
			if !iface.IsUp {
				return fmt.Errorf("restricted interface '%s' is down", restrictedIface)
			}
		}
	}

	if !foundFree {
		return fmt.Errorf("free interface '%s' not found", freeIface)
	}
	if !foundRestricted {
		return fmt.Errorf("restricted interface '%s' not found", restrictedIface)
	}

	return nil
}

// FilterUpInterfaces filters a list of interfaces to only those that are up
func FilterUpInterfaces(interfaces []NetworkInterface) []NetworkInterface {
	var up []NetworkInterface
	for _, iface := range interfaces {
		if iface.IsUp && !iface.IsLoopback {
			up = append(up, iface)
		}
	}
	return up
}

// CountUpInterfaces returns the number of interfaces that are up (excluding loopback)
func CountUpInterfaces(interfaces []NetworkInterface) int {
	count := 0
	for _, iface := range interfaces {
		if iface.IsUp && !iface.IsLoopback {
			count++
		}
	}
	return count
}
