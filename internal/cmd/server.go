package cmd

import (
	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/handlers"
	"github.com/net2share/nethopper/internal/menu"
	"github.com/spf13/cobra"
)

// NewServerRoot creates the root cobra command for nhserver.
func NewServerRoot(version, buildTime string) *cobra.Command {
	actions.RegisterServerActions()
	handlers.RegisterServerHandlers()

	root := &cobra.Command{
		Use:   "nhserver",
		Short: "Nethopper server - share internet access via reverse proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return menu.RunServerInteractive(version, buildTime)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	RegisterActionsWithRoot(root)
	return root
}
