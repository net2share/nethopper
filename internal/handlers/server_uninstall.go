package handlers

import (
	"fmt"
	"os"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/firewall"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xui"
)

func HandleServerUninstall(ctx *actions.Context) error {
	var serverCfg config.ServerConfig
	hasConfig := config.LoadJSON(config.ServerConfigPath(), &serverCfg) == nil

	if hasConfig && serverCfg.XUIMode {
		return handleXUIUninstall(ctx, &serverCfg)
	}
	return handleStandaloneUninstall(ctx)
}

func handleXUIUninstall(ctx *actions.Context, serverCfg *config.ServerConfig) error {
	keepXUI := ctx.GetBool("keep-xui")
	xuiRunning := xui.IsXUIRunning()

	beginProgress(ctx, "Uninstalling Nethopper Server (x-ui mode)")

	step, total := 1, 4
	if keepXUI {
		total = 2
	}

	// Remove x-ui entries (unless --keep-xui or x-ui is down)
	if !keepXUI {
		if !xuiRunning {
			ctx.Output.Warning("x-ui is not running, skipping x-ui cleanup")
		} else {
			user := ctx.GetString("xui-user")
			pass := ctx.GetString("xui-pass")
			if !ctx.IsInteractive && (user == "" || pass == "") {
				ctx.Output.Warning("x-ui credentials not provided, skipping x-ui cleanup")
			} else {
				// Step: Authenticate and remove from x-ui
				ctx.Output.Step(step, total, "Removing inbounds from x-ui")
				panelInfo, err := xui.ReadPanelInfo()
				if err != nil {
					ctx.Output.Warning(fmt.Sprintf("Failed to read x-ui settings: %v", err))
				} else {
					client := xui.NewClient(panelInfo)
					if err := client.Login(user, pass); err != nil {
						ctx.Output.Warning(fmt.Sprintf("Failed to authenticate: %v", err))
					} else {
						// Delete inbounds
						if err := client.DeleteInbound(serverCfg.XUISocksInboundID); err != nil {
							ctx.Output.Warning(fmt.Sprintf("Failed to delete SOCKS5 inbound: %v", err))
						}
						if err := client.DeleteInbound(serverCfg.XUITunnelInboundID); err != nil {
							ctx.Output.Warning(fmt.Sprintf("Failed to delete tunnel inbound: %v", err))
						}
						ctx.Output.Success("Inbounds removed")

						// Remove portal and routing
						step++
						ctx.Output.Step(step, total, "Removing portal and routing rules")
						xrayCfg, err := client.GetXrayConfig()
						if err != nil {
							ctx.Output.Warning(fmt.Sprintf("Failed to fetch xray config: %v", err))
						} else {
							xui.RemovePortalAndRouting(xrayCfg)
							if err := client.SaveXrayConfig(xrayCfg); err != nil {
								ctx.Output.Warning(fmt.Sprintf("Failed to save xray config: %v", err))
							}
							if err := client.RestartXray(); err != nil {
								ctx.Output.Warning(fmt.Sprintf("Failed to restart xray: %v", err))
							}
							ctx.Output.Success("Portal and routing rules removed")
						}
					}
				}
			}
		}
		step++
	}

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

	// Remove nethopper config (NOT xray binary — it belongs to x-ui)
	ctx.Output.Step(step, total, "Removing configuration")
	os.Remove(config.ServerConfigPath())
	os.Remove(config.ServerXrayConfigPath())
	os.RemoveAll(config.ServerConfigDir)
	os.RemoveAll(config.ServerStateDir)
	ctx.Output.Success("Configuration removed")

	endProgress(ctx)
	return nil
}

func handleStandaloneUninstall(ctx *actions.Context) error {
	sysMgr := service.NewSystemdManager()
	mgr := binary.NewServerManager()
	hasService := sysMgr.IsInstalled()
	hasConfig := fileExists(config.ServerConfigPath())
	hasBinary := mgr.IsInstalled(binary.XrayDef)

	if !hasService && !hasConfig && !hasBinary {
		return fmt.Errorf("nothing to uninstall")
	}

	beginProgress(ctx, "Uninstalling Nethopper Server")

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
