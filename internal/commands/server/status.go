package server

import (
	"encoding/json"
	"fmt"

	"github.com/nethopper/nethopper/internal/binary"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/service"
	"github.com/nethopper/nethopper/internal/util"
)

// StatusResult contains the status information
type StatusResult struct {
	BinaryPath       string `json:"binary_path"`
	BinaryInstalled  bool   `json:"binary_installed"`
	ConfigPath       string `json:"config_path"`
	ConfigExists     bool   `json:"config_exists"`
	ServiceName      string `json:"service_name"`
	ServiceInstalled bool   `json:"service_installed"`
	ServiceRunning   bool   `json:"service_running"`
	ServiceEnabled   bool   `json:"service_enabled"`

	// Configuration values (if config exists)
	ListenPort             uint16 `json:"listen_port,omitempty"`
	MTProtoPort            uint16 `json:"mtproto_port,omitempty"`
	VMESSPort              uint16 `json:"vmess_port,omitempty"`
	MTProtoSecret          string `json:"mtproto_secret,omitempty"`
	VMESSUuid              string `json:"vmess_uuid,omitempty"`
	FallbackHost           string `json:"fallback_host,omitempty"`
	MultiplexPerConnection int    `json:"multiplex_per_connection,omitempty"`
	LogLevel               string `json:"log_level,omitempty"`

	// Connection string
	ConnectionString string `json:"connection_string,omitempty"`
}

// StatusOptions contains options for the status command
type StatusOptions struct {
	JSON       bool
	ShowLogs   bool
	LogLines   int
	FollowLogs bool
	ServerIP   string // Override server IP for connection string
}

// Status returns the server status
func Status(opts *StatusOptions) (*StatusResult, error) {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	svcMgr := service.NewSystemdManager()

	result := &StatusResult{
		BinaryPath:       paths.BinaryPath(true),
		BinaryInstalled:  extractor.IsBinaryInstalled(true),
		ConfigPath:       paths.ConfigPath(true),
		ConfigExists:     configMgr.ConfigExists(true),
		ServiceName:      service.ServiceName,
		ServiceInstalled: svcMgr.IsInstalled(),
	}

	// Get service status
	if svcMgr.IsInstalled() {
		svcStatus, err := svcMgr.Status()
		if err == nil {
			result.ServiceRunning = svcStatus.Running
			result.ServiceEnabled = svcStatus.Enabled
		}
	}

	// Load configuration
	if result.ConfigExists {
		cfg, err := configMgr.LoadServerConfig()
		if err == nil {
			result.ListenPort = cfg.ListenPort
			result.MTProtoPort = cfg.MTProtoPort
			result.VMESSPort = cfg.VMESSPort
			result.MTProtoSecret = cfg.MTProtoSecret
			result.VMESSUuid = cfg.VMESSUuid
			result.FallbackHost = cfg.FallbackHost
			result.MultiplexPerConnection = cfg.MultiplexPerConnection
			result.LogLevel = cfg.LogLevel

			// Generate connection string
			serverIP := opts.ServerIP
			if serverIP == "" {
				serverIP = GetServerIP()
			}
			connStr := config.FromServerConfig(serverIP, cfg)
			result.ConnectionString, _ = connStr.Encode()
		}
	}

	return result, nil
}

// PrintStatus prints the status in human-readable format
func PrintStatus(result *StatusResult, printFn func(string, ...interface{}), colorOK, colorWarn, colorErr func(string) string) {
	printFn("")
	printFn("nethopper Server Status")
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

	// Service status
	if result.ServiceInstalled {
		serviceStatus := colorWarn("inactive")
		if result.ServiceRunning {
			serviceStatus = colorOK("active, running")
		}
		enabledStatus := "disabled"
		if result.ServiceEnabled {
			enabledStatus = "enabled"
		}
		printFn("Service:    %s.service (%s, %s)", result.ServiceName, serviceStatus, enabledStatus)
	} else {
		printFn("Service:    %s.service %s", result.ServiceName, colorErr("✗ Not installed"))
	}

	// Configuration values
	if result.ConfigExists && result.ListenPort > 0 {
		printFn("")
		printFn("Configuration:")
		printFn("  Listen Port:      %d", result.ListenPort)
		printFn("  MTProto Port:     %d", result.MTProtoPort)
		printFn("  Outbound VMess Port: %d", result.VMESSPort)
		printFn("  MTProto Secret:   %s", util.MaskSecret(result.MTProtoSecret))
		printFn("  VMess UUID:       %s", util.MaskUUID(result.VMESSUuid))
		printFn("  Fallback Host:    %s", result.FallbackHost)
		printFn("  Multiplex:        %d streams/connection", result.MultiplexPerConnection)
		printFn("  Log Level:        %s", result.LogLevel)
	}

	// Connection string
	if result.ConnectionString != "" {
		printFn("")
		printFn("Connection String (for freenet nodes):")
		printFn("  %s", result.ConnectionString)
	}

	printFn("")
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

// ViewLogs shows the service logs
func ViewLogs(follow bool, lines int) error {
	svcMgr := service.NewSystemdManager()
	if !svcMgr.IsInstalled() {
		return fmt.Errorf("service not installed")
	}
	return svcMgr.ViewLogs(follow, lines)
}

// GetConnectionString returns the connection string for the current configuration
func GetConnectionString(serverIP string) (string, error) {
	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)

	if !configMgr.ConfigExists(true) {
		return "", fmt.Errorf("configuration not found")
	}

	cfg, err := configMgr.LoadServerConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	if serverIP == "" {
		serverIP = GetServerIP()
	}

	connStr := config.FromServerConfig(serverIP, cfg)
	return connStr.Encode()
}
