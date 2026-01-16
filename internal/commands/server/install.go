package server

import (
	"fmt"
	"net"
	"os"
	"os/user"

	"github.com/nethopper/nethopper/internal/binary"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/firewall"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/service"
	"github.com/nethopper/nethopper/internal/util"
)

// InstallOptions contains options for the install command
type InstallOptions struct {
	ListenPort             uint16
	MTProtoPort            uint16
	VMESSPort              uint16
	MTProtoSecret          string
	FallbackHost           string
	MultiplexPerConnection int
	LogLevel               string
	SkipFirewall           bool
	NonInteractive         bool
	Verbose                bool
}

// InstallResult contains the result of the install command
type InstallResult struct {
	BinaryPath       string
	ConfigPath       string
	ServiceName      string
	ConnectionString string
	Config           *config.ServerConfig
}

// Install performs the server installation
func Install(opts *InstallOptions, printFn func(string, ...interface{})) (*InstallResult, error) {
	// Check root
	if !isRoot() {
		return nil, fmt.Errorf("server mode requires root privileges")
	}

	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	svcMgr := service.NewSystemdManager()
	fwMgr := firewall.DetectFirewall()

	// Check existing installation
	if svcMgr.IsInstalled() && configMgr.ConfigExists(true) {
		return nil, fmt.Errorf("nethopper is already installed. Use 'configure' to modify or 'uninstall' first")
	}

	// Prepare configuration
	cfg := config.ServerDefaults()

	// Set or auto-select ports
	var err error
	cfg.ListenPort, err = util.ValidateOrSelectPort(opts.ListenPort, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to select listen port: %w", err)
	}
	printFn("Listen port: %d", cfg.ListenPort)

	cfg.MTProtoPort, err = util.ValidateOrSelectPort(opts.MTProtoPort, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to select MTProto port: %w", err)
	}
	printFn("MTProto port: %d", cfg.MTProtoPort)

	cfg.VMESSPort, err = util.ValidateOrSelectPort(opts.VMESSPort, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to select Outbound VMess port: %w", err)
	}
	printFn("Outbound VMess port: %d", cfg.VMESSPort)

	// Generate or use provided secrets
	if opts.MTProtoSecret != "" {
		if !util.ValidateMTProtoSecret(opts.MTProtoSecret) {
			return nil, fmt.Errorf("invalid MTProto secret: must be 32 hex characters")
		}
		cfg.MTProtoSecret = opts.MTProtoSecret
	} else {
		cfg.MTProtoSecret, err = util.GenerateMTProtoSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to generate MTProto secret: %w", err)
		}
	}
	printFn("MTProto secret: %s", util.MaskSecret(cfg.MTProtoSecret))

	cfg.VMESSUuid = util.GenerateUUID()
	printFn("VMess UUID: %s", util.MaskUUID(cfg.VMESSUuid))

	// Set other options
	if opts.FallbackHost != "" {
		cfg.FallbackHost = opts.FallbackHost
	}
	if opts.MultiplexPerConnection > 0 {
		cfg.MultiplexPerConnection = opts.MultiplexPerConnection
	}
	if opts.LogLevel != "" {
		cfg.LogLevel = opts.LogLevel
	}

	// Extract binary
	printFn("Extracting sing-box binary...")
	extractResult, err := extractor.Extract(true)
	if err != nil {
		return nil, fmt.Errorf("failed to extract binary: %w", err)
	}
	if extractResult.Extracted {
		printFn("Binary extracted to: %s", extractResult.Path)
	} else {
		printFn("Binary already current at: %s", extractResult.Path)
	}

	// Set CAP_NET_BIND_SERVICE capability to allow binding to privileged ports (< 1024)
	// This is safer than running as root and is a standard practice for network services
	printFn("Setting network bind capability...")
	if err := binary.SetNetBindCapability(extractResult.Path); err != nil {
		printFn("Warning: failed to set CAP_NET_BIND_SERVICE capability: %v", err)
		printFn("The service may not be able to bind to ports below 1024")
	} else {
		printFn("CAP_NET_BIND_SERVICE capability set (allows binding to low ports)")
	}

	// Write configuration
	printFn("Writing configuration...")
	if err := configMgr.WriteServerConfig(&cfg); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}
	configPath := configMgr.GetConfigPath(true)
	printFn("Configuration written to: %s", configPath)

	// Create systemd service
	printFn("Creating systemd service...")
	svcConfig := &service.ServiceConfig{
		BinaryPath: extractResult.Path,
		ConfigPath: configPath,
		User:       service.ServiceUser,
		Group:      service.ServiceGroup,
	}
	if err := svcMgr.Install(svcConfig); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	printFn("Service created: %s", service.ServiceName)

	// Handle firewall
	if !opts.SkipFirewall && fwMgr.Type() != firewall.FirewallTypeNone && fwMgr.IsEnabled() {
		printFn("Firewall detected: %s", fwMgr.Type())

		ports := []uint16{cfg.ListenPort, cfg.MTProtoPort}
		if err := firewall.AllowPorts(fwMgr, ports, "tcp", "nethopper"); err != nil {
			printFn("Warning: failed to add firewall rules: %v", err)
		} else {
			printFn("Firewall rules added for ports: %d, %d", cfg.ListenPort, cfg.MTProtoPort)
		}
	}

	// Enable and start service
	printFn("Enabling and starting service...")
	if err := svcMgr.Enable(); err != nil {
		printFn("Warning: failed to enable service: %v", err)
	}
	if err := svcMgr.Start(); err != nil {
		return nil, fmt.Errorf("failed to start service: %w", err)
	}
	printFn("Service started successfully")

	// Generate connection string
	serverIP := getLocalIP()
	connStr := config.FromServerConfig(serverIP, &cfg)
	connStrEncoded, _ := connStr.Encode()

	return &InstallResult{
		BinaryPath:       extractResult.Path,
		ConfigPath:       configPath,
		ServiceName:      service.ServiceName,
		ConnectionString: connStrEncoded,
		Config:           &cfg,
	}, nil
}

// isRoot checks if the current user is root
func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}

// getLocalIP returns the local IP address (best effort)
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}

// CheckInstallation checks if net2share is installed
func CheckInstallation() (bool, bool, error) {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	svcMgr := service.NewSystemdManager()

	hasBinary := extractor.IsBinaryInstalled(true)
	hasConfig := configMgr.ConfigExists(true)
	hasService := svcMgr.IsInstalled()

	return hasBinary && hasConfig && hasService, hasConfig, nil
}

// GetServerIP prompts for or detects the server IP
func GetServerIP() string {
	// Try to detect the local IP first
	localIP := getLocalIP()

	// If running on a server with public IP, try to detect it
	// For now, just return the local IP
	return localIP
}

// EnsureRoot returns an error if not running as root
func EnsureRoot() error {
	if !isRoot() {
		if os.Getenv("SUDO_USER") != "" {
			return fmt.Errorf("please run with sudo or as root")
		}
		return fmt.Errorf("server mode requires root privileges. Run with: sudo nethopper server")
	}
	return nil
}
