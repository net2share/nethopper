package freenet

import (
	"github.com/nethopper/nethopper/internal/binary"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
)

// UninstallOptions contains options for the uninstall command
type UninstallOptions struct {
	Force      bool
	KeepConfig bool
	Verbose    bool
}

// Uninstall removes the freenet installation
func Uninstall(opts *UninstallOptions, printFn func(string, ...interface{})) error {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)

	// Remove binary
	if extractor.IsBinaryInstalled(false) {
		printFn("Removing binary...")
		if err := extractor.Remove(false); err != nil {
			printFn("Warning: failed to remove binary: %v", err)
		} else {
			printFn("Binary removed: %s", paths.BinaryPath(false))
		}
	} else {
		printFn("Binary not installed")
	}

	// Remove config (unless --keep-config)
	if !opts.KeepConfig && configMgr.ConfigExists(false) {
		printFn("Removing configuration...")
		if err := configMgr.RemoveConfig(false); err != nil {
			printFn("Warning: failed to remove config: %v", err)
		} else {
			printFn("Configuration removed")
		}
	} else if opts.KeepConfig {
		printFn("Keeping configuration at: %s", paths.ConfigPath(false))
	}

	printFn("Uninstall complete")
	return nil
}

// CheckUninstallable checks if there's anything to uninstall
func CheckUninstallable() (bool, error) {
	paths := platform.NewPathProvider()
	extractor := binary.NewExtractor(paths)
	configMgr := config.NewManager(paths)

	hasBinary := extractor.IsBinaryInstalled(false)
	hasConfig := configMgr.ConfigExists(false)

	return hasBinary || hasConfig, nil
}
