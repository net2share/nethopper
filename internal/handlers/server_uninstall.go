package handlers

import (
	"fmt"
	"os"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/firewall"
	"github.com/net2share/nethopper/internal/service"
)

func HandleServerUninstall(ctx *actions.Context) error {
	// Check if there's anything to uninstall
	sysMgr := service.NewSystemdManager()
	mgr := binary.NewServerManager()
	hasService := sysMgr.IsInstalled()
	hasConfig := fileExists(config.ServerConfigPath())
	hasBinary := mgr.IsInstalled(binary.XrayDef)

	if !hasService && !hasConfig && !hasBinary {
		return fmt.Errorf("nothing to uninstall")
	}

	beginProgress(ctx, "Uninstalling Nethopper Server")

	// Load config to get ports for firewall cleanup
	var serverCfg config.ServerConfig
	_ = config.LoadJSON(config.ServerConfigPath(), &serverCfg)

	step, total := 1, 4

	// Stop and remove systemd service
	ctx.Output.Step(step, total, "Removing systemd service")
	if hasService {
		if err := sysMgr.Uninstall(); err != nil {
			ctx.Output.Warning(fmt.Sprintf("Failed to remove service: %v", err))
		} else {
			ctx.Output.Success("Service removed")
		}
	} else {
		ctx.Output.Info("No service installed")
	}
	step++

	// Remove firewall rules
	ctx.Output.Step(step, total, "Removing firewall rules")
	if serverCfg.SocksPort > 0 {
		fm := firewall.DetectFirewall()
		ports := []uint16{uint16(serverCfg.SocksPort), uint16(serverCfg.TunnelPort)}
		if err := firewall.RemovePorts(fm, ports, "tcp"); err != nil {
			ctx.Output.Warning(fmt.Sprintf("Failed to remove firewall rules: %v", err))
		} else {
			ctx.Output.Success("Firewall rules removed")
		}
	}
	step++

	// Remove config
	ctx.Output.Step(step, total, "Removing configuration")
	os.Remove(config.ServerConfigPath())
	os.Remove(config.ServerXrayConfigPath())
	os.RemoveAll(config.ServerConfigDir)
	os.RemoveAll(config.ServerStateDir)
	ctx.Output.Success("Configuration removed")
	step++

	// Remove binary
	ctx.Output.Step(step, total, "Removing xray binary")
	if hasBinary {
		if err := mgr.Remove(binary.XrayDef); err != nil {
			ctx.Output.Warning(fmt.Sprintf("Failed to remove binary: %v", err))
		} else {
			ctx.Output.Success("Binary removed")
		}
	} else {
		ctx.Output.Info("No binary installed")
	}

	endProgress(ctx)
	return nil
}
