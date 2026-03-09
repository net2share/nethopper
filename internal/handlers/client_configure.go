package handlers

import (
	"fmt"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/xray"
)

func HandleClientConfigure(ctx *actions.Context) error {
	connStr := ctx.GetString("connection")
	if connStr == "" {
		return fmt.Errorf("connection string is required")
	}

	cs, err := config.DecodeConnectionString(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	freeIface := ctx.GetString("free-interface")
	if freeIface == "" {
		return fmt.Errorf("free interface is required")
	}

	restrictedIface := ctx.GetString("restricted-interface")
	if restrictedIface == "" {
		return fmt.Errorf("restricted interface is required")
	}

	if freeIface == restrictedIface {
		return fmt.Errorf("free and restricted interfaces must be different")
	}

	clientCfg := cs.ToClientConfig(freeIface, restrictedIface)
	if err := config.SaveJSON(config.ClientConfigPath(), clientCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Generate xray config
	xrayCfg, err := xray.GenerateClientConfig(clientCfg)
	if err != nil {
		return fmt.Errorf("failed to generate xray config: %w", err)
	}
	if err := writeFile(config.ClientXrayConfigPath(), xrayCfg); err != nil {
		return fmt.Errorf("failed to write xray config: %w", err)
	}

	ctx.Output.Success("Client configured")
	ctx.Output.Info(fmt.Sprintf("Server: %s:%d", cs.ServerIP, cs.TunnelPort))
	ctx.Output.Info(fmt.Sprintf("Free interface: %s", freeIface))
	ctx.Output.Info(fmt.Sprintf("Restricted interface: %s", restrictedIface))
	return nil
}
