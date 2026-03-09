package menu

import (
	"os"

	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
	"github.com/net2share/nethopper/internal/service"
)

// serverState checks what's installed on the server.
type serverState struct {
	hasService bool
	hasConfig  bool
	hasBinary  bool
}

func getServerState() serverState {
	sysMgr := service.NewSystemdManager()
	mgr := binary.NewServerManager()
	return serverState{
		hasService: sysMgr.IsInstalled(),
		hasConfig:  fileExists(config.ServerConfigPath()),
		hasBinary:  mgr.IsInstalled(binary.XrayDef),
	}
}

func (s serverState) isInstalled() bool {
	return s.hasService || s.hasConfig || s.hasBinary
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
