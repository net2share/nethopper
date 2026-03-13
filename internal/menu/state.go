package menu

import (
	"os"

	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/service"
	"github.com/net2share/nethopper/internal/xui"
)

// serverState checks what's installed on the server.
type serverState struct {
	hasService  bool
	hasConfig   bool
	hasBinary   bool
	xuiDetected bool
	xuiMode     bool
}

func getServerState() serverState {
	sysMgr := service.NewSystemdManager()
	mgr := binary.NewServerManager()

	state := serverState{
		hasService:  sysMgr.IsInstalled(),
		hasConfig:   fileExists(config.ServerConfigPath()),
		hasBinary:   mgr.IsInstalled(binary.XrayDef),
		xuiDetected: xui.IsXUIRunning(),
	}

	if state.hasConfig {
		var cfg config.ServerConfig
		if config.LoadJSON(config.ServerConfigPath(), &cfg) == nil {
			state.xuiMode = cfg.XUIMode
		}
	}

	return state
}

func (s serverState) isInstalled() bool {
	return s.hasService || s.hasConfig || s.hasBinary || s.xuiMode
}

// clientState checks what's installed on the client.
type clientState struct {
	hasConfig bool
	hasBinary bool
}

func getClientState() clientState {
	mgr := binary.NewClientManager(config.ClientBinDir())
	return clientState{
		hasConfig: fileExists(config.ClientConfigPath()),
		hasBinary: mgr.IsInstalled(binary.XrayDef),
	}
}

func (s clientState) isInstalled() bool {
	return s.hasConfig || s.hasBinary
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
