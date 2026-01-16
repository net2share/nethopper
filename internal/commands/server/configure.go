package server

import (
	"fmt"

	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/service"
	"github.com/nethopper/nethopper/internal/util"
)

// ConfigureOptions contains options for the configure command
type ConfigureOptions struct {
	ListenPort              uint16
	MTProtoPort             uint16
	VMESSPort               uint16
	MTProtoSecret           string
	RegenerateMTProtoSecret bool // Force regeneration of MTProto secret
	FallbackHost            string
	MultiplexPerConnection  int
	LogLevel                string
	CustomConfig            []byte // If provided, use this instead of generating
	NonInteractive          bool
	Verbose                 bool
}

// Configure modifies the server configuration
func Configure(opts *ConfigureOptions, printFn func(string, ...interface{})) (*config.ServerConfig, error) {
	// Check root
	if err := EnsureRoot(); err != nil {
		return nil, err
	}

	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)
	svcMgr := service.NewSystemdManager()

	// Handle custom config
	if len(opts.CustomConfig) > 0 {
		printFn("Using custom configuration...")
		if err := configMgr.WriteCustomConfig(true, opts.CustomConfig); err != nil {
			return nil, fmt.Errorf("failed to write custom config: %w", err)
		}

		// Restart service if running
		if svcMgr.IsInstalled() {
			status, _ := svcMgr.Status()
			if status != nil && status.Running {
				printFn("Restarting service...")
				if err := svcMgr.Restart(); err != nil {
					return nil, fmt.Errorf("failed to restart service: %w", err)
				}
			}
		}

		return nil, nil
	}

	// Load existing config or create new
	var cfg *config.ServerConfig
	var err error

	if configMgr.ConfigExists(true) {
		printFn("Loading existing configuration...")
		cfg, err = configMgr.LoadServerConfig()
		if err != nil {
			printFn("Warning: failed to load existing config, using defaults")
			defaults := config.ServerDefaults()
			cfg = &defaults
		}
	} else {
		defaults := config.ServerDefaults()
		cfg = &defaults
	}

	// Update only provided values
	if opts.ListenPort > 0 {
		if !util.IsPortAvailable(opts.ListenPort) && opts.ListenPort != cfg.ListenPort {
			return nil, fmt.Errorf("port %d is not available", opts.ListenPort)
		}
		cfg.ListenPort = opts.ListenPort
	} else if cfg.ListenPort == 0 {
		cfg.ListenPort, err = util.FindAvailablePort(5)
		if err != nil {
			return nil, fmt.Errorf("failed to find available port: %w", err)
		}
	}

	if opts.MTProtoPort > 0 {
		if !util.IsPortAvailable(opts.MTProtoPort) && opts.MTProtoPort != cfg.MTProtoPort {
			return nil, fmt.Errorf("port %d is not available", opts.MTProtoPort)
		}
		cfg.MTProtoPort = opts.MTProtoPort
	} else if cfg.MTProtoPort == 0 {
		cfg.MTProtoPort, err = util.FindAvailablePort(5)
		if err != nil {
			return nil, fmt.Errorf("failed to find available port: %w", err)
		}
	}

	if opts.VMESSPort > 0 {
		if !util.IsPortAvailable(opts.VMESSPort) && opts.VMESSPort != cfg.VMESSPort {
			return nil, fmt.Errorf("port %d is not available", opts.VMESSPort)
		}
		cfg.VMESSPort = opts.VMESSPort
	} else if cfg.VMESSPort == 0 {
		cfg.VMESSPort, err = util.FindAvailablePort(5)
		if err != nil {
			return nil, fmt.Errorf("failed to find available port: %w", err)
		}
	}

	if opts.RegenerateMTProtoSecret || cfg.MTProtoSecret == "" {
		cfg.MTProtoSecret, err = util.GenerateMTProtoSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secret: %w", err)
		}
		printFn("Generated new MTProto secret")
	} else if opts.MTProtoSecret != "" {
		if !util.ValidateMTProtoSecret(opts.MTProtoSecret) {
			return nil, fmt.Errorf("invalid MTProto secret: must be 32 hex characters")
		}
		cfg.MTProtoSecret = opts.MTProtoSecret
	}

	if cfg.VMESSUuid == "" {
		cfg.VMESSUuid = util.GenerateUUID()
	}

	if opts.FallbackHost != "" {
		cfg.FallbackHost = opts.FallbackHost
	}

	if opts.MultiplexPerConnection > 0 {
		cfg.MultiplexPerConnection = opts.MultiplexPerConnection
	}

	if opts.LogLevel != "" {
		cfg.LogLevel = opts.LogLevel
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Write configuration
	printFn("Writing configuration...")
	if err := configMgr.WriteServerConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}
	printFn("Configuration updated: %s", configMgr.GetConfigPath(true))

	// Restart service if running
	if svcMgr.IsInstalled() {
		status, _ := svcMgr.Status()
		if status != nil && status.Running {
			printFn("Restarting service...")
			if err := svcMgr.Restart(); err != nil {
				return nil, fmt.Errorf("failed to restart service: %w", err)
			}
			printFn("Service restarted")
		}
	}

	return cfg, nil
}
