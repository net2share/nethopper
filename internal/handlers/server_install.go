package handlers

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/firewall"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xray"
	"github.com/net2share/nethopper/internal/xui"
)

func HandleServerInstall(ctx *actions.Context) error {
	xuiMode := false
	if ctx.IsInteractive {
		xuiMode = ctx.GetString("install-mode") == "xui"
	} else {
		xuiMode = xui.IsXUIRunning() && !ctx.GetBool("standalone")
	}

	if xuiMode {
		return handleXUIInstall(ctx)
	}
	return handleStandaloneInstall(ctx)
}

func handleXUIInstall(ctx *actions.Context) error {
	// Validate credentials
	user := ctx.GetString("xui-user")
	pass := ctx.GetString("xui-pass")
	if !ctx.IsInteractive && (user == "" || pass == "") {
		return fmt.Errorf("x-ui detected: --xui-user and --xui-pass are required")
	}

	beginProgress(ctx, "Installing via x-ui")

	// Step 1: Read panel info
	ctx.Output.Step(1, 7, "Reading x-ui panel configuration")
	panelInfo, err := xui.ReadPanelInfo()
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to read x-ui settings: %w", err))
	}
	ctx.Output.Success(fmt.Sprintf("Panel found at %s", panelInfo.BaseURL))

	// Step 2: Authenticate
	ctx.Output.Step(2, 7, "Authenticating with x-ui panel")
	client := xui.NewClient(panelInfo)
	if err := client.Login(user, pass); err != nil {
		return failProgress(ctx, fmt.Errorf("authentication failed: %w", err))
	}
	ctx.Output.Success("Authenticated")

	// Step 3: Select/validate ports
	ctx.Output.Step(3, 7, "Selecting ports")
	socksPort := ctx.GetInt("socks-port")
	tunnelPort := ctx.GetInt("tunnel-port")
	if socksPort == 0 {
		if p, err := randomAvailablePort(); err == nil {
			socksPort = p
		} else {
			return failProgress(ctx, fmt.Errorf("failed to find available port: %w", err))
		}
	} else if !isPortAvailable(socksPort) {
		return failProgress(ctx, fmt.Errorf("SOCKS5 port %d is not available", socksPort))
	}
	if tunnelPort == 0 {
		if p, err := randomAvailablePort(); err == nil {
			tunnelPort = p
		} else {
			return failProgress(ctx, fmt.Errorf("failed to find available port: %w", err))
		}
	} else if !isPortAvailable(tunnelPort) {
		return failProgress(ctx, fmt.Errorf("tunnel port %d is not available", tunnelPort))
	}
	ctx.Output.Success(fmt.Sprintf("SOCKS5: %d, Tunnel: %d", socksPort, tunnelPort))

	// Step 4: Generate UUID + add inbounds
	ctx.Output.Step(4, 7, "Creating inbounds in x-ui")
	uuid, err := generateUUID()
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to generate UUID: %w", err))
	}
	socksID, socksTag, err := client.AddSocksInbound(socksPort)
	if err != nil {
		return failProgress(ctx, err)
	}
	tunnelID, tunnelTag, err := client.AddVLESSInbound(tunnelPort, uuid)
	if err != nil {
		// Clean up the socks inbound we just created
		client.DeleteInbound(socksID)
		return failProgress(ctx, err)
	}
	ctx.Output.Success(fmt.Sprintf("Inbounds created (SOCKS5: #%d, Tunnel: #%d)", socksID, tunnelID))

	// Step 5: Update xray template (portal + routing)
	ctx.Output.Step(5, 7, "Configuring reverse portal")
	xrayCfg, err := client.GetXrayConfig()
	if err != nil {
		return failProgress(ctx, err)
	}
	xui.AddPortalAndRouting(xrayCfg, socksTag, tunnelTag)
	if err := client.SaveXrayConfig(xrayCfg); err != nil {
		return failProgress(ctx, err)
	}
	if err := client.RestartXray(); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Failed to restart xray: %v", err))
	}
	ctx.Output.Success("Reverse portal configured")

	// Step 6: Configure firewall
	ctx.Output.Step(6, 7, "Configuring firewall")
	fm := firewall.DetectFirewall()
	ports := []uint16{uint16(socksPort), uint16(tunnelPort)}
	if err := firewall.AllowPorts(fm, ports, "tcp", "nethopper"); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Firewall configuration failed: %v", err))
	} else {
		ctx.Output.Success(fmt.Sprintf("Firewall configured (%s)", fm.Type()))
	}

	// Step 7: Save config
	ctx.Output.Step(7, 7, "Saving configuration")
	serverCfg := &config.ServerConfig{
		SocksPort:          socksPort,
		TunnelPort:         tunnelPort,
		UUID:               uuid,
		XUIMode:            true,
		XUISocksInboundID:  socksID,
		XUITunnelInboundID: tunnelID,
		XUISocksTag:        socksTag,
		XUITunnelTag:       tunnelTag,
		XUIPortalTag:       xui.NHPortalTag,
	}
	if err := config.SaveJSON(config.ServerConfigPath(), serverCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to save config: %w", err))
	}
	ctx.Output.Success("Configuration saved")

	endProgress(ctx)
	return nil
}

func handleStandaloneInstall(ctx *actions.Context) error {
	beginProgress(ctx, "Installing Nethopper Server")

	// Step 1: Download/ensure xray binary
	ctx.Output.Step(1, 5, "Ensuring xray binary is available")
	mgr := binary.NewServerManager()
	xrayPath, err := mgr.EnsureInstalled(binary.XrayDef, nil)
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to install xray: %w", err))
	}

	// If binary is outside the standard server dir (e.g. env override), copy it
	standardPath := filepath.Join(config.ServerBinDir, binary.XrayDef.Name)
	if xrayPath != standardPath {
		if err := copyFile(xrayPath, standardPath); err != nil {
			return failProgress(ctx, fmt.Errorf("failed to copy xray binary: %w", err))
		}
		os.Chmod(standardPath, 0755)
		xrayPath = standardPath
	}
	ctx.Output.Success(fmt.Sprintf("Xray binary: %s", xrayPath))

	// Step 2: Generate UUID and config
	ctx.Output.Step(2, 5, "Generating server configuration")
	uuid, err := generateUUID()
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to generate UUID: %w", err))
	}

	socksPort := ctx.GetInt("socks-port")
	tunnelPort := ctx.GetInt("tunnel-port")
	if socksPort == 0 {
		if p, err := randomAvailablePort(); err == nil {
			socksPort = p
		} else {
			socksPort = 1080
		}
	}
	if tunnelPort == 0 {
		if p, err := randomAvailablePort(); err == nil {
			tunnelPort = p
		} else {
			tunnelPort = 2083
		}
	}

	serverCfg := &config.ServerConfig{
		SocksPort:  socksPort,
		TunnelPort: tunnelPort,
		UUID:       uuid,
	}

	if err := config.SaveJSON(config.ServerConfigPath(), serverCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to save config: %w", err))
	}

	// Generate xray config
	xrayCfg, err := xray.GenerateServerConfig(serverCfg)
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to generate xray config: %w", err))
	}
	if err := writeFile(config.ServerXrayConfigPath(), xrayCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to write xray config: %w", err))
	}
	exec.Command("chown", "-R", "nethopper:nethopper", config.ServerConfigDir).Run()
	ctx.Output.Success("Configuration saved")

	// Step 3: Configure firewall
	ctx.Output.Step(3, 5, "Configuring firewall")
	fm := firewall.DetectFirewall()
	fwPorts := []uint16{uint16(serverCfg.SocksPort), uint16(serverCfg.TunnelPort)}
	if err := firewall.AllowPorts(fm, fwPorts, "tcp", "nethopper"); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Firewall configuration failed: %v", err))
	} else {
		ctx.Output.Success(fmt.Sprintf("Firewall configured (%s)", fm.Type()))
	}

	// Step 4: Install systemd service
	ctx.Output.Step(4, 5, "Installing systemd service")
	sysMgr := service.NewSystemdManager()
	svcCfg := &service.ServiceConfig{
		BinaryPath: xrayPath,
		ConfigPath: config.ServerXrayConfigPath(),
	}
	if err := sysMgr.Install(svcCfg); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to install service: %w", err))
	}
	ctx.Output.Success("Systemd service installed")

	// Step 5: Start service
	ctx.Output.Step(5, 5, "Starting service")
	if err := sysMgr.Enable(); err != nil {
		ctx.Output.Warning(fmt.Sprintf("Failed to enable service: %v", err))
	}
	if err := sysMgr.Start(); err != nil {
		return failProgress(ctx, fmt.Errorf("failed to start service: %w", err))
	}
	ctx.Output.Success("Service started")

	endProgress(ctx)
	return nil
}

func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return "", err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
