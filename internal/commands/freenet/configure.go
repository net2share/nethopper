package freenet

import (
	"fmt"
	"os"

	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
)

// ConfigureOptions contains options for the configure command
type ConfigureOptions struct {
	// From connection string
	ConnectionString string

	// Manual configuration
	ServerIP            string
	ServerPort          uint16
	VMESSPort           uint16
	VMESSUuid           string
	FreeInterface       string
	RestrictedInterface string
	MaxConnections      int
	LogLevel            string

	// Custom config
	CustomConfigPath string
	CustomConfigJSON []byte

	NonInteractive bool
	Verbose        bool
}

// Configure sets up the freenet configuration
func Configure(opts *ConfigureOptions, printFn func(string, ...interface{})) (*config.FreenetConfig, error) {
	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)

	// Handle custom config file
	if opts.CustomConfigPath != "" {
		printFn("Reading custom config from: %s", opts.CustomConfigPath)
		data, err := os.ReadFile(opts.CustomConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := configMgr.WriteCustomConfig(false, data); err != nil {
			return nil, fmt.Errorf("failed to write config: %w", err)
		}
		printFn("Configuration written to: %s", configMgr.GetConfigPath(false))
		return nil, nil
	}

	// Handle custom config JSON
	if len(opts.CustomConfigJSON) > 0 {
		printFn("Using provided JSON configuration...")
		if err := configMgr.WriteCustomConfig(false, opts.CustomConfigJSON); err != nil {
			return nil, fmt.Errorf("failed to write config: %w", err)
		}
		printFn("Configuration written to: %s", configMgr.GetConfigPath(false))
		return nil, nil
	}

	// Start with defaults or existing config
	var cfg *config.FreenetConfig

	if configMgr.ConfigExists(false) {
		printFn("Loading existing configuration...")
		var err error
		cfg, err = configMgr.LoadFreenetConfig()
		if err != nil {
			printFn("Warning: failed to load existing config, using defaults")
			defaults := config.FreenetDefaults()
			cfg = &defaults
		}
	} else {
		defaults := config.FreenetDefaults()
		cfg = &defaults
	}

	// Handle connection string
	if opts.ConnectionString != "" {
		printFn("Parsing connection string...")
		connStr, err := config.DecodeConnectionString(opts.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("invalid connection string: %w", err)
		}
		printFn("Decoded: Server %s:%d, VMess port %d", connStr.ServerIP, connStr.Port, connStr.VMESSPort)

		cfg.ServerIP = connStr.ServerIP
		cfg.ServerPort = connStr.Port
		cfg.VMESSPort = connStr.VMESSPort
		cfg.VMESSUuid = connStr.UUID
	} else {
		// Manual configuration - update only provided values
		if opts.ServerIP != "" {
			cfg.ServerIP = opts.ServerIP
		}
		if opts.ServerPort > 0 {
			cfg.ServerPort = opts.ServerPort
		}
		if opts.VMESSPort > 0 {
			cfg.VMESSPort = opts.VMESSPort
		}
		if opts.VMESSUuid != "" {
			cfg.VMESSUuid = opts.VMESSUuid
		}
	}

	// Interface configuration
	if opts.FreeInterface != "" {
		cfg.FreeInterface = opts.FreeInterface
	}
	if opts.RestrictedInterface != "" {
		cfg.RestrictedInterface = opts.RestrictedInterface
	}

	// Other options
	if opts.MaxConnections > 0 {
		cfg.MaxConnections = opts.MaxConnections
	}
	if opts.LogLevel != "" {
		cfg.LogLevel = opts.LogLevel
	}

	// Validate required fields for non-interactive mode
	if opts.NonInteractive {
		if cfg.ServerIP == "" {
			return nil, fmt.Errorf("server-ip is required")
		}
		if cfg.ServerPort == 0 {
			return nil, fmt.Errorf("server-port is required")
		}
		if cfg.VMESSPort == 0 {
			return nil, fmt.Errorf("vmess-port is required")
		}
		if cfg.VMESSUuid == "" {
			return nil, fmt.Errorf("vmess-uuid is required")
		}
		if cfg.FreeInterface == "" {
			return nil, fmt.Errorf("free-iface is required")
		}
		if cfg.RestrictedInterface == "" {
			return nil, fmt.Errorf("restricted-iface is required")
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Validate interfaces if provided
	if cfg.FreeInterface != "" && cfg.RestrictedInterface != "" {
		netMgr := platform.NewNetworkManager()
		interfaces, err := netMgr.ListInterfaces()
		if err != nil {
			printFn("Warning: could not validate interfaces: %v", err)
		} else {
			if err := platform.ValidateInterfaceSelection(interfaces, cfg.FreeInterface, cfg.RestrictedInterface); err != nil {
				return nil, fmt.Errorf("interface validation failed: %w", err)
			}
			printFn("Interfaces validated")
		}
	}

	// Write configuration
	printFn("Writing configuration...")
	if err := configMgr.WriteFreenetConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}
	printFn("Configuration written to: %s", configMgr.GetConfigPath(false))

	return cfg, nil
}

// ListAvailableInterfaces returns a list of available network interfaces
func ListAvailableInterfaces() ([]platform.NetworkInterface, error) {
	netMgr := platform.NewNetworkManager()
	return netMgr.ListUpInterfaces()
}
