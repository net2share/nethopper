package handlers

import (
	"context"
	"fmt"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/network"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xray"
)

func HandleClientRun(ctx *actions.Context) error {
	// Load config
	var clientCfg config.ClientConfig
	if err := config.LoadJSON(config.ClientConfigPath(), &clientCfg); err != nil {
		return fmt.Errorf("client not configured: %w", err)
	}
	if err := clientCfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Validate interfaces
	if err := network.ValidateInterface(clientCfg.FreeInterface); err != nil {
		return fmt.Errorf("free interface error: %w", err)
	}
	if err := network.ValidateInterface(clientCfg.RestrictedInterface); err != nil {
		return fmt.Errorf("restricted interface error: %w", err)
	}

	// Regenerate xray config with current interface names
	xrayCfg, err := xray.GenerateClientConfig(&clientCfg)
	if err != nil {
		return fmt.Errorf("failed to generate xray config: %w", err)
	}
	if err := writeFile(config.ClientXrayConfigPath(), xrayCfg); err != nil {
		return fmt.Errorf("failed to write xray config: %w", err)
	}

	// Ensure xray binary
	mgr := binary.NewClientManager(config.ClientBinDir())
	xrayPath, err := mgr.EnsureInstalled(binary.XrayDef, nil)
	if err != nil {
		return fmt.Errorf("failed to ensure xray binary: %w", err)
	}

	ctx.Output.Info(fmt.Sprintf("Connecting to %s:%d", clientCfg.ServerIP, clientCfg.TunnelPort))
	ctx.Output.Info(fmt.Sprintf("Free: %s | Restricted: %s",
		clientCfg.FreeInterface, clientCfg.RestrictedInterface))
	ctx.Output.Info("Press Ctrl+C to stop")

	// Run xray in foreground
	pm := service.NewProcessManager(xrayPath, config.ClientXrayConfigPath())
	return pm.Run(context.Background())
}
