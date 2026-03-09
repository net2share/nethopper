package handlers

import "github.com/net2share/nethopper/internal/actions"

// RegisterServerHandlers binds handlers to server actions.
func RegisterServerHandlers() {
	actions.SetHandler(actions.ActionServerInstall, HandleServerInstall)
	actions.SetHandler(actions.ActionServerConfigure, HandleServerConfigure)
	actions.SetHandler(actions.ActionServerStatus, HandleServerStatus)
	actions.SetHandler(actions.ActionServerUninstall, HandleServerUninstall)
}

// RegisterClientHandlers binds handlers to client actions.
func RegisterClientHandlers() {
	actions.SetHandler(actions.ActionClientInstall, HandleClientInstall)
	actions.SetHandler(actions.ActionClientConfigure, HandleClientConfigure)
	actions.SetHandler(actions.ActionClientRun, HandleClientRun)
	actions.SetHandler(actions.ActionClientStatus, HandleClientStatus)
	actions.SetHandler(actions.ActionClientUninstall, HandleClientUninstall)
}
