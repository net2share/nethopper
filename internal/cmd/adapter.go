package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/net2share/go-corelib/osdetect"
	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/handlers"
	"github.com/spf13/cobra"
)

// BuildCobraCommand builds a Cobra command from an action.
func BuildCobraCommand(action *actions.Action) *cobra.Command {
	cmd := &cobra.Command{
		Use:    action.Use,
		Short:  action.Short,
		Long:   action.Long,
		Hidden: action.Hidden,
	}

	// Add flags for inputs
	for _, input := range action.Inputs {
		if input.InteractiveOnly {
			continue
		}
		switch input.Type {
		case actions.InputTypeText, actions.InputTypePassword, actions.InputTypeSelect:
			if input.ShortFlag != 0 {
				cmd.Flags().StringP(input.Name, string(input.ShortFlag), input.Default, input.Label)
			} else {
				cmd.Flags().String(input.Name, input.Default, input.Label)
			}
		case actions.InputTypeNumber:
			defaultVal := 0
			if input.Default != "" {
				if v, err := strconv.Atoi(input.Default); err == nil {
					defaultVal = v
				}
			}
			if input.ShortFlag != 0 {
				cmd.Flags().IntP(input.Name, string(input.ShortFlag), defaultVal, input.Label)
			} else {
				cmd.Flags().Int(input.Name, defaultVal, input.Label)
			}
		case actions.InputTypeBool:
			cmd.Flags().Bool(input.Name, false, input.Label)
		}
	}

	// Handle confirmation flag
	if action.Confirm != nil && action.Confirm.ForceFlag != "" {
		cmd.Flags().BoolP(action.Confirm.ForceFlag, "f", false, "Skip confirmation")
	}

	if action.IsSubmenu {
		return cmd
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if action.RequiresRoot {
			if err := osdetect.RequireRoot(); err != nil {
				return err
			}
		}

		ctx := &actions.Context{
			Ctx:           context.Background(),
			Args:          args,
			Values:        make(map[string]interface{}),
			Output:        handlers.NewTUIOutput(),
			IsInteractive: false,
		}

		// Collect values from flags
		for _, input := range action.Inputs {
			if input.InteractiveOnly {
				continue
			}
			switch input.Type {
			case actions.InputTypeText, actions.InputTypePassword, actions.InputTypeSelect:
				val, _ := cmd.Flags().GetString(input.Name)
				ctx.Values[input.Name] = val
			case actions.InputTypeNumber:
				val, _ := cmd.Flags().GetInt(input.Name)
				ctx.Values[input.Name] = val
			case actions.InputTypeBool:
				val, _ := cmd.Flags().GetBool(input.Name)
				ctx.Values[input.Name] = val
			}
		}

		// Handle confirmation
		if action.Confirm != nil && action.Confirm.ForceFlag != "" {
			force, _ := cmd.Flags().GetBool(action.Confirm.ForceFlag)
			if !force {
				return fmt.Errorf("%s\n\nUse --force to confirm", action.Confirm.Message)
			}
		}

		if action.Handler == nil {
			return fmt.Errorf("no handler for action %s", action.ID)
		}

		return action.Handler(ctx)
	}

	return cmd
}

// BuildAllCommands builds all Cobra commands from registered actions.
func BuildAllCommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, action := range actions.TopLevel() {
		cmd := BuildCobraCommand(action)
		for _, child := range actions.GetChildren(action.ID) {
			cmd.AddCommand(BuildCobraCommand(child))
		}
		commands = append(commands, cmd)
	}
	return commands
}

// RegisterActionsWithRoot adds all action-based commands to a root command.
func RegisterActionsWithRoot(root *cobra.Command) {
	for _, cmd := range BuildAllCommands() {
		root.AddCommand(cmd)
	}
}
