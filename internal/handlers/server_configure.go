package handlers

import (
	"fmt"
	"os/exec"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xray"
)

func HandleServerConfigure(ctx *actions.Context) error {
	beginProgress(ctx, "Configuring Server")

	// Load existing config
	var serverCfg config.ServerConfig
	if err := config.LoadJSON(config.ServerConfigPath(), &serverCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to load config (run install first): %w", err))
	}

	// Apply new values
	if port := ctx.GetInt("socks-port"); port > 0 {
		serverCfg.SocksPort = port
	}
	if port := ctx.GetInt("tunnel-port"); port > 0 {
		serverCfg.TunnelPort = port
	}

	// Save updated config
	if err := config.SaveJSON(config.ServerConfigPath(), &serverCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to save config: %w", err))
	}

	// Regenerate xray config
	xrayCfg, err := xray.GenerateServerConfig(&serverCfg)
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to generate xray config: %w", err))
	}
	if err := writeFile(config.ServerXrayConfigPath(), xrayCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to write xray config: %w", err))
	}
	exec.Command("chown", "-R", "nethopper:nethopper", config.ServerConfigDir).Run()
	ctx.Output.Success("Configuration updated")

	// Restart service
	sysMgr := service.NewSystemdManager()
	if sysMgr.IsInstalled() {
		ctx.Output.Status("Restarting service")
		if err := sysMgr.Restart(); err != nil {
			ctx.Output.Warning(fmt.Sprintf("Failed to restart service: %v", err))
		} else {
			ctx.Output.Success("Service restarted")
		}
	}

	endProgress(ctx)
	return nil
}
