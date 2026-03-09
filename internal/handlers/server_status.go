package handlers

import (
	"fmt"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/service"
)

func HandleServerStatus(ctx *actions.Context) error {
	var rows []actions.InfoRow

	// Binary status
	mgr := binary.NewServerManager()
	xrayPath, err := mgr.ResolvePath(binary.XrayDef)
	if err != nil {
		rows = append(rows, actions.InfoRow{Key: "Xray", Value: "not installed"})
	} else {
		rows = append(rows, actions.InfoRow{Key: "Xray", Value: xrayPath})
	}

	// Config status
	var serverCfg config.ServerConfig
	var connStr string
	if err := config.LoadJSON(config.ServerConfigPath(), &serverCfg); err == nil {
		rows = append(rows,
			actions.InfoRow{Key: "SOCKS5 Port", Value: fmt.Sprintf("%d", serverCfg.SocksPort)},
			actions.InfoRow{Key: "Tunnel Port", Value: fmt.Sprintf("%d", serverCfg.TunnelPort)},
			actions.InfoRow{Key: "UUID", Value: serverCfg.UUID},
		)

		// Connection string
		serverIP := detectServerIP()
		cs := config.NewConnectionString(serverIP, &serverCfg)
		if encoded, err := cs.Encode(); err == nil {
			connStr = encoded
		}
	} else {
		rows = append(rows, actions.InfoRow{Key: "Config", Value: "not configured"})
	}

	// Service status
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

	if connStr != "" {
		rows = append(rows, actions.InfoRow{Key: "Connection String", Value: connStr})
	}

	return ctx.Output.ShowInfo(actions.InfoConfig{
		Title:    "Server Status",
		CopyText: connStr,
		Sections: []actions.InfoSection{{Rows: rows}},
	})
}
