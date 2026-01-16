package server

import (
	"github.com/nethopper/nethopper/internal/binary"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/firewall"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/service"
)

// UninstallOptions contains options for the uninstall command
type UninstallOptions struct {
	Force          bool
	KeepConfig     bool
	RemoveFirewall bool
	Verbose        bool
}

// Uninstall removes the server installation
func Uninstall(opts *UninstallOptions, printFn func(string, ...interface{})) error {
	// Check root
	if err := EnsureRoot(); err != nil {
		return err
	}

	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	svcMgr := service.NewSystemdManager()
	fwMgr := firewall.DetectFirewall()

	// Load config to get ports for firewall cleanup
	var cfg *config.ServerConfig
	if configMgr.ConfigExists(true) {
		cfg, _ = configMgr.LoadServerConfig()
	}

	// Stop and remove service
	if svcMgr.IsInstalled() {
		printFn("Stopping service...")
		if err := svcMgr.Stop(); err != nil {
			printFn("Warning: failed to stop service: %v", err)
		}

		printFn("Removing systemd service...")
		if err := svcMgr.Uninstall(); err != nil {
			printFn("Warning: failed to remove service: %v", err)
		} else {
			printFn("Service removed")
		}
	} else {
		printFn("Service not installed")
	}

	// Remove binary
	if extractor.IsBinaryInstalled(true) {
		printFn("Removing binary...")
		if err := extractor.Remove(true); err != nil {
			printFn("Warning: failed to remove binary: %v", err)
		} else {
			printFn("Binary removed: %s", paths.BinaryPath(true))
		}
	} else {
		printFn("Binary not installed")
	}

	// Remove config (unless --keep-config)
	if !opts.KeepConfig && configMgr.ConfigExists(true) {
		printFn("Removing configuration...")
		if err := configMgr.RemoveConfig(true); err != nil {
			printFn("Warning: failed to remove config: %v", err)
		} else {
			printFn("Configuration removed")
		}
	} else if opts.KeepConfig {
		printFn("Keeping configuration at: %s", paths.ConfigPath(true))
	}

	// Remove firewall rules
	if opts.RemoveFirewall && cfg != nil && fwMgr.Type() != firewall.FirewallTypeNone {
		printFn("Removing firewall rules...")
		ports := []uint16{cfg.ListenPort, cfg.MTProtoPort}
		if err := firewall.RemovePorts(fwMgr, ports, "tcp"); err != nil {
			printFn("Warning: failed to remove firewall rules: %v", err)
		} else {
			printFn("Firewall rules removed")
		}
	}

	printFn("Uninstall complete")
	return nil
}

// CheckUninstallable checks if there's anything to uninstall
func CheckUninstallable() (bool, error) {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)
	svcMgr := service.NewSystemdManager()

	hasBinary := extractor.IsBinaryInstalled(true)
	hasConfig := configMgr.ConfigExists(true)
	hasService := svcMgr.IsInstalled()

	return hasBinary || hasConfig || hasService, nil
}
