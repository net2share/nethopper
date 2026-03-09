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
)

func HandleServerInstall(ctx *actions.Context) error {
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

	serverCfg := &config.ServerConfig{
		SocksPort:  1080,
		TunnelPort: 2083,
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
	// Set ownership so the nethopper service user can read config
	exec.Command("chown", "-R", "nethopper:nethopper", config.ServerConfigDir).Run()

	ctx.Output.Success("Configuration saved")

	// Step 3: Configure firewall
	ctx.Output.Step(3, 5, "Configuring firewall")
	fm := firewall.DetectFirewall()
	ports := []uint16{uint16(serverCfg.SocksPort), uint16(serverCfg.TunnelPort)}
	if err := firewall.AllowPorts(fm, ports, "tcp", "nethopper"); err != nil {
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
