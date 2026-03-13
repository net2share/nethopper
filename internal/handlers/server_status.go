package handlers

import (
	"fmt"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xui"
)

func HandleServerStatus(ctx *actions.Context) error {
	var rows []actions.InfoRow
	var connStr string

	var serverCfg config.ServerConfig
	if err := config.LoadJSON(config.ServerConfigPath(), &serverCfg); err == nil {
		if serverCfg.XUIMode {
			rows = append(rows, actions.InfoRow{Key: "Mode", Value: "x-ui integration"})

			// x-ui service status
			if xui.IsXUIRunning() {
				rows = append(rows, actions.InfoRow{Key: "x-ui Service", Value: "active"})
			} else {
				rows = append(rows, actions.InfoRow{Key: "x-ui Service", Value: "inactive"})
			}

			rows = append(rows,
				actions.InfoRow{Key: "SOCKS5 Inbound", Value: fmt.Sprintf("#%d (%s)", serverCfg.XUISocksInboundID, serverCfg.XUISocksTag)},
				actions.InfoRow{Key: "Tunnel Inbound", Value: fmt.Sprintf("#%d (%s)", serverCfg.XUITunnelInboundID, serverCfg.XUITunnelTag)},
			)
		} else {
			rows = append(rows, actions.InfoRow{Key: "Mode", Value: "standalone"})

			// Binary status
			mgr := binary.NewServerManager()
			xrayPath, err := mgr.ResolvePath(binary.XrayDef)
			if err != nil {
				rows = append(rows, actions.InfoRow{Key: "Xray", Value: "not installed"})
			} else {
				rows = append(rows, actions.InfoRow{Key: "Xray", Value: xrayPath})
			}
		}

		// Common fields
		rows = append(rows,
			actions.InfoRow{Key: "SOCKS5 Port", Value: fmt.Sprintf("%d", serverCfg.SocksPort)},
			actions.InfoRow{Key: "Tunnel Port", Value: fmt.Sprintf("%d", serverCfg.TunnelPort)},
			actions.InfoRow{Key: "UUID", Value: serverCfg.UUID},
		)

		// Service status (standalone only)
		if !serverCfg.XUIMode {
			sysMgr := service.NewSystemdManager()
			if sysMgr.IsInstalled() {
				status, err := sysMgr.Status()
				if err == nil {
					rows = append(rows,
						actions.InfoRow{Key: "Service", Value: status.Active},
						actions.InfoRow{Key: "Enabled", Value: fmt.Sprintf("%v", status.Enabled)},
					)
				}
			} else {
				rows = append(rows, actions.InfoRow{Key: "Service", Value: "not installed"})
			}
		}

		// Connection string
		serverIP := detectServerIP()
		cs := config.NewConnectionString(serverIP, &serverCfg)
		if encoded, err := cs.Encode(); err == nil {
			connStr = encoded
		}
	} else {
		rows = append(rows, actions.InfoRow{Key: "Config", Value: "not configured"})
	}

	if connStr != "" {
		rows = append(rows, actions.InfoRow{Key: "Connection String", Value: connStr})
	}

	return ctx.Output.ShowInfo(actions.InfoConfig{
		Title:    "Server Status",
		CopyText: connStr,
		Sections: []actions.InfoSection{{Rows: rows}},
	})
}
