package menu

import (
	"context"
	"fmt"

	"github.com/net2share/go-corelib/tui"
	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/handlers"
	"github.com/net2share/nethopper/internal/network"
)

// RunAction executes an action in interactive mode.
// If inline is true, output goes directly to the terminal (no fullscreen TUI).
func RunAction(actionID string, inline ...bool) error {
	action := actions.Get(actionID)
	if action == nil {
		return fmt.Errorf("unknown action: %s", actionID)
	}

	useInline := len(inline) > 0 && inline[0]
	output := handlers.NewInteractiveTUIOutput()
	if useInline {
		output = handlers.NewTUIOutput()
	}

	ctx := &actions.Context{
		Ctx:           context.Background(),
		Values:        make(map[string]interface{}),
		Output:        output,
		IsInteractive: true,
	}

	// Collect inputs interactively
	if err := collectInputs(ctx, action); err != nil {
		return err
	}

	// Handle confirmation
	if action.Confirm != nil {
		confirm, err := tui.RunConfirm(tui.ConfirmConfig{
			Title:   action.Confirm.Message,
			Default: !action.Confirm.DefaultNo,
		})
		if err != nil {
			return err
		}
		if !confirm {
			return errCancelled
		}
	}

	if action.Handler == nil {
		return fmt.Errorf("no handler for action %s", action.ID)
	}

	return action.Handler(ctx)
}

// collectInputs collects action inputs interactively via TUI forms.
func collectInputs(ctx *actions.Context, action *actions.Action) error {
	for _, input := range action.Inputs {
		if input.ShowIf != nil && !input.ShowIf(ctx) {
			continue
		}

		var value interface{}

		switch input.Type {
		case actions.InputTypeText, actions.InputTypePassword:
			defaultVal := input.Default
			if input.DefaultFunc != nil {
				defaultVal = input.DefaultFunc(ctx)
			}

			description := input.Description
			if defaultVal != "" && description != "" {
				description = fmt.Sprintf("%s (default: %s)", description, defaultVal)
			}

			var validationErr error
			for {
				desc := description
				if validationErr != nil {
					desc = fmt.Sprintf("%s\n⚠ %s", description, validationErr.Error())
				}

				val, confirmed, err := tui.RunInput(tui.InputConfig{
					Title:       input.Label,
					Description: desc,
					Placeholder: input.Placeholder,
					Value:       defaultVal,
					Password:    input.Type == actions.InputTypePassword,
				})
				if err != nil {
					return err
				}
				if !confirmed {
					return errCancelled
				}
				if val == "" && defaultVal != "" {
					val = defaultVal
				}
				if input.Required && val == "" {
					validationErr = fmt.Errorf("%s is required", input.Label)
					continue
				}
				if input.Validate != nil && val != "" {
					if validationErr = input.Validate(val); validationErr != nil {
						continue
					}
				}
				value = val
				break
			}

		case actions.InputTypeNumber:
			defaultVal := input.Default
			if input.DefaultFunc != nil {
				defaultVal = input.DefaultFunc(ctx)
			}

			description := input.Description
			if defaultVal != "" {
				description = fmt.Sprintf("%s (default: %s)", description, defaultVal)
			}

			var validationErr error
			for {
				desc := description
				if validationErr != nil {
					desc = fmt.Sprintf("%s\n⚠ %s", description, validationErr.Error())
				}

				val, confirmed, err := tui.RunInput(tui.InputConfig{
					Title:       input.Label,
					Description: desc,
					Value:       defaultVal,
				})
				if err != nil {
					return err
				}
				if !confirmed {
					return errCancelled
				}
				if val == "" && defaultVal != "" {
					val = defaultVal
				}
				if input.Validate != nil && val != "" {
					if validationErr = input.Validate(val); validationErr != nil {
						continue
					}
				}
				var intVal int
				fmt.Sscanf(val, "%d", &intVal)
				value = intVal
				break
			}

		case actions.InputTypeSelect:
			var tuiOptions []tui.MenuOption
			options := input.Options
			if input.OptionsFunc != nil {
				options = input.OptionsFunc(ctx)
			}

			// Special handling: interface selection
			if input.Name == "free-interface" || input.Name == "restricted-interface" {
				ifaces, err := network.ListInterfaces()
				if err != nil {
					return fmt.Errorf("failed to list interfaces: %w", err)
				}
				// Filter out already-selected interface
				var exclude string
				if input.Name == "restricted-interface" {
					exclude, _ = ctx.Values["free-interface"].(string)
				} else if input.Name == "free-interface" {
					exclude, _ = ctx.Values["restricted-interface"].(string)
				}
				for _, iface := range ifaces {
					if iface.Name == exclude {
						continue
					}
					tuiOptions = append(tuiOptions, tui.MenuOption{
						Label: fmt.Sprintf("%s (%s)", iface.Name, iface.IP),
						Value: iface.Name,
					})
				}
			} else {
				for _, opt := range options {
					tuiOptions = append(tuiOptions, tui.MenuOption{
						Label: opt.Label,
						Value: opt.Value,
					})
				}
			}

			if !input.Required {
				tuiOptions = append(tuiOptions, tui.MenuOption{Label: "Skip", Value: ""})
			}

			val, err := tui.RunMenu(tui.MenuConfig{
				Title:       input.Label,
				Description: input.Description,
				Options:     tuiOptions,
			})
			if err != nil {
				return err
			}
			if val == "" && input.Required {
				return errCancelled
			}
			value = val

		case actions.InputTypeBool:
			continue
		}

		ctx.Values[input.Name] = value
	}
	return nil
}
