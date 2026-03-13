package actions

import (
	"fmt"
	"strconv"
)

var portValidator = func(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("invalid port number")
	}
	return nil
}

// RegisterServerActions registers all server-side actions.
func RegisterServerActions() {
	Register(&Action{
		ID:           ActionServerInstall,
		Use:          "install",
		Short:        "Install xray and configure server",
		RequiresRoot: true,
		Inputs: []InputField{
			{
				Name:            "install-mode",
				Label:           "Installation mode",
				Type:            InputTypeSelect,
				InteractiveOnly: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetBool("xui-detected")
				},
				Options: []SelectOption{
					{Label: "Configure via x-ui panel", Value: "xui"},
					{Label: "Install standalone", Value: "standalone"},
				},
			},
			{
				Name:  "standalone",
				Label: "Install standalone (skip x-ui integration)",
				Type:  InputTypeBool,
			},
			{
				Name:     "xui-user",
				Label:    "x-ui panel username",
				Type:     InputTypeText,
				Required: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetString("install-mode") == "xui"
				},
			},
			{
				Name:     "xui-pass",
				Label:    "x-ui panel password",
				Type:     InputTypePassword,
				Required: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetString("install-mode") == "xui"
				},
			},
			{
				Name:        "socks-port",
				Label:       "SOCKS5 port",
				Description: "Port for SOCKS5 inbound",
				Type:        InputTypeNumber,
				Validate:    portValidator,
			},
			{
				Name:        "tunnel-port",
				Label:       "Tunnel port",
				Description: "Port for VLESS tunnel inbound",
				Type:        InputTypeNumber,
				Validate:    portValidator,
			},
		},
	})

	Register(&Action{
		ID:           ActionServerConfigure,
		Use:          "configure",
		Short:        "Configure server settings",
		RequiresRoot: true,
		Inputs: []InputField{
			{
				Name:     "xui-user",
				Label:    "x-ui panel username",
				Type:     InputTypeText,
				Required: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetBool("xui-mode")
				},
			},
			{
				Name:     "xui-pass",
				Label:    "x-ui panel password",
				Type:     InputTypePassword,
				Required: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetBool("xui-mode")
				},
			},
			{
				Name:        "socks-port",
				Label:       "SOCKS5 port",
				Description: "Port for SOCKS5 inbound",
				Type:        InputTypeNumber,
				Validate:    portValidator,
			},
			{
				Name:        "tunnel-port",
				Label:       "Tunnel port",
				Description: "Port for VLESS tunnel inbound",
				Type:        InputTypeNumber,
				Validate:    portValidator,
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
		Inputs: []InputField{
			{
				Name:     "xui-user",
				Label:    "x-ui panel username",
				Type:     InputTypeText,
				Required: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetBool("xui-mode")
				},
			},
			{
				Name:     "xui-pass",
				Label:    "x-ui panel password",
				Type:     InputTypePassword,
				Required: true,
				ShowIf: func(ctx *Context) bool {
					return ctx.GetBool("xui-mode")
				},
			},
			{
				Name:  "keep-xui",
				Label: "Keep x-ui configuration (only remove nethopper config)",
				Type:  InputTypeBool,
			},
		},
		Confirm: &ConfirmConfig{
			Message:   "Remove nethopper server configuration?",
			DefaultNo: true,
			ForceFlag: "force",
		},
	})
}
