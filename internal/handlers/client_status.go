package handlers

import (
	"fmt"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
)

func HandleClientStatus(ctx *actions.Context) error {
	var rows []actions.InfoRow

	// Binary status
	mgr := binary.NewClientManager(config.ClientBinDir())
	if xrayPath, err := mgr.ResolvePath(binary.XrayDef); err == nil {
		rows = append(rows, actions.InfoRow{Key: "Xray", Value: xrayPath})
	} else {
		rows = append(rows, actions.InfoRow{Key: "Xray", Value: "not installed"})
	}

	// Config status
	var clientCfg config.ClientConfig
	if err := config.LoadJSON(config.ClientConfigPath(), &clientCfg); err == nil {
		rows = append(rows,
			actions.InfoRow{Key: "Server", Value: fmt.Sprintf("%s:%d", clientCfg.ServerIP, clientCfg.TunnelPort)},
			actions.InfoRow{Key: "Free Interface", Value: clientCfg.FreeInterface},
			actions.InfoRow{Key: "Restricted Interface", Value: clientCfg.RestrictedInterface},
			actions.InfoRow{Key: "UUID", Value: clientCfg.UUID},
		)
	} else {
		rows = append(rows, actions.InfoRow{Key: "Config", Value: "not configured"})
	}

	return ctx.Output.ShowInfo(actions.InfoConfig{
		Title:    "Client Status",
		Sections: []actions.InfoSection{{Rows: rows}},
	})
}
