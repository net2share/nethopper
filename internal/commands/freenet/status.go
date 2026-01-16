package freenet

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nethopper/nethopper/internal/binary"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
)

// InterfaceStatus represents the status of a network interface
type InterfaceStatus struct {
	Name      string `json:"name"`
	Found     bool   `json:"found"`
	Up        bool   `json:"up"`
	IPv4      string `json:"ipv4,omitempty"`
	Type      string `json:"type,omitempty"`
}

// StatusResult contains the status information
type StatusResult struct {
	BinaryPath      string `json:"binary_path"`
	BinaryInstalled bool   `json:"binary_installed"`
	ConfigPath      string `json:"config_path"`
	ConfigExists    bool   `json:"config_exists"`

	// Configuration values (if config exists)
	ServerIP            string `json:"server_ip,omitempty"`
	ServerPort          uint16 `json:"server_port,omitempty"`
	VMESSPort           uint16 `json:"vmess_port,omitempty"`
	VMESSUuid           string `json:"vmess_uuid,omitempty"`
	MaxConnections      int    `json:"max_connections,omitempty"`
	LogLevel            string `json:"log_level,omitempty"`

	// Interface status
	FreeInterface       *InterfaceStatus `json:"free_interface,omitempty"`
	RestrictedInterface *InterfaceStatus `json:"restricted_interface,omitempty"`

	// Connectivity
	ServerReachable bool          `json:"server_reachable,omitempty"`
	PingLatency     time.Duration `json:"ping_latency,omitempty"`
}

// StatusOptions contains options for the status command
type StatusOptions struct {
	JSON bool
	Ping bool
}

// Status returns the freenet node status
func Status(opts *StatusOptions) (*StatusResult, error) {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	netMgr := platform.NewNetworkManager()

	result := &StatusResult{
		BinaryPath:      paths.BinaryPath(false),
		BinaryInstalled: extractor.IsBinaryInstalled(false),
		ConfigPath:      paths.ConfigPath(false),
		ConfigExists:    configMgr.ConfigExists(false),
	}

	// Load configuration
	if result.ConfigExists {
		cfg, err := configMgr.LoadFreenetConfig()
		if err == nil {
			result.ServerIP = cfg.ServerIP
			result.ServerPort = cfg.ServerPort
			result.VMESSPort = cfg.VMESSPort
			result.VMESSUuid = cfg.VMESSUuid
			result.MaxConnections = cfg.MaxConnections
			result.LogLevel = cfg.LogLevel

			// Check interfaces
			result.FreeInterface = checkInterface(netMgr, cfg.FreeInterface)
			result.RestrictedInterface = checkInterface(netMgr, cfg.RestrictedInterface)

			// Ping test if requested
			if opts.Ping && result.RestrictedInterface != nil && result.RestrictedInterface.Up {
				latency, err := netMgr.PingTest(cfg.ServerIP, cfg.RestrictedInterface, 5*time.Second)
				if err == nil {
					result.ServerReachable = true
					result.PingLatency = latency
				}
			}
		}
	}

	return result, nil
}

// checkInterface checks the status of a network interface
func checkInterface(netMgr platform.NetworkManager, name string) *InterfaceStatus {
	if name == "" {
		return nil
	}

	status := &InterfaceStatus{
		Name:  name,
		Found: false,
		Up:    false,
	}

	iface, err := netMgr.GetInterface(name)
	if err != nil {
		return status
	}

	status.Found = true
	status.Up = iface.IsUp

	if len(iface.IPv4Addrs) > 0 {
		status.IPv4 = iface.IPv4Addrs[0]
	}

	if iface.IsWireless {
		status.Type = "Wireless"
	} else if iface.IsEthernet {
		status.Type = "Ethernet"
	} else {
		status.Type = "Unknown"
	}

	return status
}

// PrintStatus prints the status in human-readable format
func PrintStatus(result *StatusResult, printFn func(string, ...interface{}), colorOK, colorWarn, colorErr func(string) string) {
	printFn("")
	printFn("nethopper Freenet Node Status")
	printFn("═══════════════════════════════════════")
	printFn("")

	// Binary status
	binaryStatus := colorErr("✗ Not installed")
	if result.BinaryInstalled {
		binaryStatus = colorOK("✓ Installed")
	}
	printFn("Binary:     %s %s", result.BinaryPath, binaryStatus)

	// Config status
	configStatus := colorErr("✗ Not found")
	if result.ConfigExists {
		configStatus = colorOK("✓ Found")
	}
	printFn("Config:     %s %s", result.ConfigPath, configStatus)

	// Configuration values
	if result.ConfigExists && result.ServerIP != "" {
		printFn("")
		printFn("Configuration:")
		printFn("  Server:           %s:%d", result.ServerIP, result.ServerPort)
		printFn("  VMess Port:       %d", result.VMESSPort)
		printFn("  Max Connections:  %d", result.MaxConnections)
		printFn("  Log Level:        %s", result.LogLevel)
	}

	// Interface status
	if result.FreeInterface != nil || result.RestrictedInterface != nil {
		printFn("")
		printFn("Interfaces:")

		if result.FreeInterface != nil {
			status := formatInterfaceStatus(result.FreeInterface, colorOK, colorWarn, colorErr)
			printFn("  Free:        %s", status)
		}

		if result.RestrictedInterface != nil {
			status := formatInterfaceStatus(result.RestrictedInterface, colorOK, colorWarn, colorErr)
			printFn("  Restricted:  %s", status)
		}
	}

	// Connectivity
	if result.ServerReachable {
		printFn("")
		printFn("Server Connectivity:")
		printFn("  Ping via restricted interface: %s %s",
			result.PingLatency.Round(time.Millisecond).String(),
			colorOK("✓"))
	} else if result.FreeInterface != nil && result.RestrictedInterface != nil {
		printFn("")
		printFn("Server Connectivity:")
		printFn("  %s (use --ping to test)", colorWarn("Not tested"))
	}

	printFn("")
}

// formatInterfaceStatus formats an interface status for display
func formatInterfaceStatus(iface *InterfaceStatus, colorOK, colorWarn, colorErr func(string) string) string {
	if !iface.Found {
		return fmt.Sprintf("%s %s", iface.Name, colorErr("✗ NOT FOUND"))
	}

	status := colorOK("✓ UP")
	if !iface.Up {
		status = colorWarn("⚠ DOWN")
	}

	ipStr := ""
	if iface.IPv4 != "" {
		ipStr = fmt.Sprintf("(%s)", iface.IPv4)
	}

	typeStr := ""
	if iface.Type != "" {
		typeStr = fmt.Sprintf(" - %s", iface.Type)
	}

	return fmt.Sprintf("%s %s%s %s", iface.Name, ipStr, typeStr, status)
}

// PrintStatusJSON prints the status in JSON format
func PrintStatusJSON(result *StatusResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
