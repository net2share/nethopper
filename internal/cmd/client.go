package cmd

import (
	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/handlers"
	"github.com/net2share/nethopper/internal/menu"
	"github.com/spf13/cobra"
)

// NewClientRoot creates the root cobra command for nhclient.
func NewClientRoot(version, buildTime string) *cobra.Command {
	actions.RegisterClientActions()
	handlers.RegisterClientHandlers()

	root := &cobra.Command{
		Use:   "nhclient",
		Short: "Nethopper client - connect to server and provide free internet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return menu.RunClientInteractive(version, buildTime)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	RegisterActionsWithRoot(root)
	return root
}
