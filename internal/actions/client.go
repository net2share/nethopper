package actions

// RegisterClientActions registers all client-side actions.
func RegisterClientActions() {
	Register(&Action{
		ID:    ActionClientInstall,
		Use:   "install",
		Short: "Download xray binary",
	})

	Register(&Action{
		ID:    ActionClientConfigure,
		Use:   "configure",
		Short: "Configure client connection",
		Inputs: []InputField{
			{
				Name:        "connection",
				Label:       "Connection string",
				Description: "Server connection string (nh://...)",
				Type:        InputTypeText,
				Required:    true,
				Placeholder: "nh://...",
				ShortFlag:   'c',
			},
			{
				Name:        "free-interface",
				Label:       "Free internet interface",
				Description: "Network interface with free internet access",
				Type:        InputTypeSelect,
				Required:    true,
				ShortFlag:   'f',
			},
			{
				Name:        "restricted-interface",
				Label:       "Restricted internet interface",
				Description: "Network interface with restricted internet access (tunnel goes through this)",
				Type:        InputTypeSelect,
				Required:    true,
				ShortFlag:   'r',
			},
		},
	})

	Register(&Action{
		ID:    ActionClientRun,
		Use:   "run",
		Short: "Start xray bridge in foreground",
	})

	Register(&Action{
		ID:    ActionClientStatus,
		Use:   "status",
		Short: "Show client configuration and status",
	})

	Register(&Action{
		ID:    ActionClientUninstall,
		Use:   "uninstall",
		Short: "Remove client configuration and binary",
		Confirm: &ConfirmConfig{
			Message:   "Remove client config and xray binary?",
			DefaultNo: true,
			ForceFlag: "force",
		},
	})
}
