package actions

import (
	"fmt"
	"strconv"
)

// RegisterServerActions registers all server-side actions.
func RegisterServerActions() {
	Register(&Action{
		ID:           ActionServerInstall,
		Use:          "install",
		Short:        "Install xray and configure server",
		RequiresRoot: true,
	})

	Register(&Action{
		ID:           ActionServerConfigure,
		Use:          "configure",
		Short:        "Configure server settings",
		RequiresRoot: true,
		Inputs: []InputField{
			{
				Name:        "socks-port",
				Label:       "SOCKS5 port",
				Description: "Port for SOCKS5 inbound",
				Type:        InputTypeNumber,
				Default:     "1080",
				Validate: func(value string) error {
					n, err := strconv.Atoi(value)
					if err != nil || n < 1 || n > 65535 {
						return fmt.Errorf("invalid port number")
					}
					return nil
				},
			},
			{
				Name:        "tunnel-port",
				Label:       "Tunnel port",
				Description: "Port for VLESS tunnel inbound",
				Type:        InputTypeNumber,
				Default:     "2083",
				Validate: func(value string) error {
					n, err := strconv.Atoi(value)
					if err != nil || n < 1 || n > 65535 {
						return fmt.Errorf("invalid port number")
					}
					return nil
				},
			},
		},
	})

	Register(&Action{
		ID:    ActionServerStatus,
		Use:   "status",
		Short: "Show server status and connection string",
	})

	Register(&Action{
		ID:           ActionServerUninstall,
		Use:          "uninstall",
		Short:        "Remove server installation",
		RequiresRoot: true,
		Confirm: &ConfirmConfig{
			Message:   "Remove xray, config, and systemd service?",
			DefaultNo: true,
			ForceFlag: "force",
		},
	})
}
