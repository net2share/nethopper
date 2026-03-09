package handlers

import (
	"fmt"
	"os"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
)

func HandleClientUninstall(ctx *actions.Context) error {
	// Check if there's anything to uninstall
	mgr := binary.NewClientManager(config.ClientBinDir())
	hasConfig := fileExists(config.ClientConfigPath())
	hasBinary := mgr.IsInstalled(binary.XrayDef)

	if !hasConfig && !hasBinary {
		return fmt.Errorf("nothing to uninstall")
	}

	beginProgress(ctx, "Uninstalling Nethopper Client")

	step, total := 1, countTrue(hasConfig, hasBinary)

	if hasConfig {
		ctx.Output.Step(step, total, "Removing configuration")
		os.Remove(config.ClientConfigPath())
		os.Remove(config.ClientXrayConfigPath())
		os.RemoveAll(config.ClientConfigDir())
		ctx.Output.Success("Configuration removed")
		step++
	}

	if hasBinary {
		ctx.Output.Step(step, total, "Removing xray binary")
		if err := mgr.Remove(binary.XrayDef); err != nil {
			ctx.Output.Warning(fmt.Sprintf("Failed to remove binary: %v", err))
		} else {
			ctx.Output.Success("Binary removed")
		}
	}

	endProgress(ctx)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func countTrue(vals ...bool) int {
	n := 0
	for _, v := range vals {
		if v {
			n++
		}
	}
	return n
}
