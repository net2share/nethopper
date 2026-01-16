package freenet

import (
	"context"
	"fmt"

	"github.com/nethopper/nethopper/internal/binary"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/service"
)

// RunOptions contains options for the run command
type RunOptions struct {
	Verbose bool
}

// Run starts sing-box with the freenet configuration
func Run(opts *RunOptions, printFn func(string, ...interface{})) error {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	netMgr := platform.NewNetworkManager()

	// Check if config exists
	if !configMgr.ConfigExists(false) {
		return fmt.Errorf("configuration not found. Run 'nethopper freenet configure' first")
	}

	// Load and validate config
	cfg, err := configMgr.LoadFreenetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Check interfaces
	interfaces, err := netMgr.ListInterfaces()
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	if err := platform.ValidateInterfaceSelection(interfaces, cfg.FreeInterface, cfg.RestrictedInterface); err != nil {
		return fmt.Errorf("interface validation failed: %w", err)
	}
	printFn("Interfaces validated: free=%s, restricted=%s", cfg.FreeInterface, cfg.RestrictedInterface)

	// Extract binary if needed
	printFn("Checking sing-box binary...")
	extractResult, err := extractor.Extract(false)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	if extractResult.Extracted {
		printFn("Binary extracted to: %s", extractResult.Path)
	} else {
		printFn("Binary ready at: %s", extractResult.Path)
	}

	// Check binary
	if err := service.CheckBinary(extractResult.Path); err != nil {
		return fmt.Errorf("binary check failed: %w", err)
	}

	// Validate config with sing-box
	configPath := configMgr.GetConfigPath(false)
	if err := service.ValidateConfig(extractResult.Path, configPath); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	printFn("Configuration validated")

	// Start sing-box
	printFn("")
	printFn("Starting sing-box...")
	printFn("Server: %s:%d", cfg.ServerIP, cfg.ServerPort)
	printFn("Press Ctrl+C to stop")
	printFn("")

	// Run in foreground
	procMgr := service.NewProcessManager(extractResult.Path, configPath)
	if err := procMgr.Run(context.Background()); err != nil {
		return fmt.Errorf("sing-box exited with error: %w", err)
	}

	printFn("\nStopped")
	return nil
}

// CheckRunnable checks if the freenet node can be run
func CheckRunnable() error {
	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)
	netMgr := platform.NewNetworkManager()

	// Check config
	if !configMgr.ConfigExists(false) {
		return fmt.Errorf("configuration not found")
	}

	// Load config
	cfg, err := configMgr.LoadFreenetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check interfaces
	interfaces, err := netMgr.ListInterfaces()
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	if err := platform.ValidateInterfaceSelection(interfaces, cfg.FreeInterface, cfg.RestrictedInterface); err != nil {
		return err
	}

	// Check interface count
	upInterfaces := platform.FilterUpInterfaces(interfaces)
	if len(upInterfaces) < 2 {
		return fmt.Errorf("at least 2 network interfaces must be up, found %d", len(upInterfaces))
	}

	return nil
}
