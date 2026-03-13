package handlers

import (
	"fmt"
	"os/exec"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/firewall"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xray"
	"github.com/net2share/nethopper/internal/xui"
)

func HandleServerConfigure(ctx *actions.Context) error {
	var serverCfg config.ServerConfig
	if err := config.LoadJSON(config.ServerConfigPath(), &serverCfg); err != nil {
		return fmt.Errorf("failed to load config (run install first): %w", err)
	}

	if serverCfg.XUIMode {
		return handleXUIConfigure(ctx, &serverCfg)
	}
	return handleStandaloneConfigure(ctx, &serverCfg)
}

func handleXUIConfigure(ctx *actions.Context, serverCfg *config.ServerConfig) error {
	// Verify x-ui is running
	if !xui.IsXUIRunning() {
		return fmt.Errorf("x-ui is not running (required for x-ui mode configuration)")
	}

	user := ctx.GetString("xui-user")
	pass := ctx.GetString("xui-pass")
	if !ctx.IsInteractive && (user == "" || pass == "") {
		return fmt.Errorf("x-ui mode: --xui-user and --xui-pass are required")
	}

	beginProgress(ctx, "Configuring via x-ui")

	// Step 1: Read panel info and authenticate
	ctx.Output.Step(1, 5, "Authenticating with x-ui panel")
	panelInfo, err := xui.ReadPanelInfo()
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to read x-ui settings: %w", err))
	}
	client := xui.NewClient(panelInfo)
	if err := client.Login(user, pass); err != nil {
		return failProgress(ctx, fmt.Errorf("authentication failed: %w", err))
	}
	ctx.Output.Success("Authenticated")

	// Step 2: Get new port values
	ctx.Output.Step(2, 5, "Updating configuration")
	newSocksPort := ctx.GetInt("socks-port")
	newTunnelPort := ctx.GetInt("tunnel-port")
	if newSocksPort == 0 {
		newSocksPort = serverCfg.SocksPort
	}
	if newTunnelPort == 0 {
		newTunnelPort = serverCfg.TunnelPort
	}

	portsChanged := newSocksPort != serverCfg.SocksPort || newTunnelPort != serverCfg.TunnelPort

	if !portsChanged {
		ctx.Output.Info("No changes detected")
		endProgress(ctx)
		return nil
	}

	// Step 3: Delete old inbounds and create new ones
	ctx.Output.Step(3, 5, "Updating inbounds in x-ui")
	if err := client.DeleteInbound(serverCfg.XUISocksInboundID); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Failed to delete old SOCKS5 inbound: %v", err))
	}
	if err := client.DeleteInbound(serverCfg.XUITunnelInboundID); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Failed to delete old tunnel inbound: %v", err))
	}

	socksID, socksTag, err := client.AddSocksInbound(newSocksPort)
	if err != nil {
		return failProgress(ctx, err)
	}
	tunnelID, tunnelTag, err := client.AddVLESSInbound(newTunnelPort, serverCfg.UUID)
	if err != nil {
		return failProgress(ctx, err)
	}
	ctx.Output.Success(fmt.Sprintf("Inbounds updated (SOCKS5: #%d, Tunnel: #%d)", socksID, tunnelID))

	// Step 4: Update routing rules
	ctx.Output.Step(4, 5, "Updating routing rules")
	xrayCfg, err := client.GetXrayConfig()
	if err != nil {
		return failProgress(ctx, err)
	}
	xui.UpdateRoutingTags(xrayCfg, serverCfg.XUISocksTag, serverCfg.XUITunnelTag, socksTag, tunnelTag)
	if err := client.SaveXrayConfig(xrayCfg); err != nil {
		return failProgress(ctx, err)
	}
	if err := client.RestartXray(); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Failed to restart xray: %v", err))
	}
	ctx.Output.Success("Routing rules updated")

	// Step 5: Update firewall and save config
	ctx.Output.Step(5, 5, "Saving configuration")
	// Update firewall: remove old, add new
	fm := firewall.DetectFirewall()
	oldPorts := []uint16{uint16(serverCfg.SocksPort), uint16(serverCfg.TunnelPort)}
	firewall.RemovePorts(fm, oldPorts, "tcp")
	newPorts := []uint16{uint16(newSocksPort), uint16(newTunnelPort)}
	firewall.AllowPorts(fm, newPorts, "tcp", "nethopper")

	serverCfg.SocksPort = newSocksPort
	serverCfg.TunnelPort = newTunnelPort
	serverCfg.XUISocksInboundID = socksID
	serverCfg.XUITunnelInboundID = tunnelID
	serverCfg.XUISocksTag = socksTag
	serverCfg.XUITunnelTag = tunnelTag

	if err := config.SaveJSON(config.ServerConfigPath(), serverCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to save config: %w", err))
	}
	ctx.Output.Success("Configuration updated")

	endProgress(ctx)
	return nil
}

func handleStandaloneConfigure(ctx *actions.Context, serverCfg *config.ServerConfig) error {
	beginProgress(ctx, "Configuring Server")

	// Apply new values
	if port := ctx.GetInt("socks-port"); port > 0 {
		serverCfg.SocksPort = port
	}
	if port := ctx.GetInt("tunnel-port"); port > 0 {
		serverCfg.TunnelPort = port
	}

	// Save updated config
	if err := config.SaveJSON(config.ServerConfigPath(), serverCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to save config: %w", err))
	}

	// Regenerate xray config
	xrayCfg, err := xray.GenerateServerConfig(serverCfg)
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
