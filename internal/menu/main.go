// Package menu provides the interactive TUI menus for nethopper.
package menu

import (
	"errors"

	"github.com/net2share/go-corelib/tui"
	"github.com/net2share/nethopper/internal/actions"
)

var errCancelled = errors.New("cancelled")

const nethopperBanner = `
    _   __     __  __  __
   / | / /__  / /_/ / / /___  ____  ____  ___  _____
  /  |/ / _ \/ __/ /_/ / __ \/ __ \/ __ \/ _ \/ ___/
 / /|  /  __/ /_/ __  / /_/ / /_/ / /_/ /  __/ /
/_/ |_/\___/\__/_/ /_/\____/ .___/ .___/\___/_/
                          /_/   /_/
`

// RunServerInteractive shows the server interactive menu.
func RunServerInteractive(version, buildTime string) error {
	tui.SetAppInfo("nhserver", version, buildTime)
	tui.BeginSession()
	defer tui.EndSession()

	tui.PrintBanner(tui.BannerConfig{
		AppName:   "Nethopper Server",
		Version:   version,
		BuildTime: buildTime,
		ASCII:     nethopperBanner,
	})

	return runServerMenu()
}

// RunClientInteractive shows the client interactive menu.
func RunClientInteractive(version, buildTime string) error {
	tui.SetAppInfo("nhclient", version, buildTime)
	tui.BeginSession()
	defer tui.EndSession()

	tui.PrintBanner(tui.BannerConfig{
		AppName:   "Nethopper Client",
		Version:   version,
		BuildTime: buildTime,
		ASCII:     nethopperBanner,
	})

	return runClientMenu()
}

// isInlineAction returns true for actions whose output should be printed
// directly on the terminal instead of inside a fullscreen TUI component.
func isInlineAction(actionID string) bool {
	return actionID == actions.ActionServerStatus
}

func runServerMenu() error {
	for {
		state := getServerState()
		var options []tui.MenuOption

		if !state.hasConfig {
			options = append(options, tui.MenuOption{Label: "Install", Value: actions.ActionServerInstall})
		}
		if state.hasConfig {
			options = append(options, tui.MenuOption{Label: "Configure", Value: actions.ActionServerConfigure})
		}
		if state.isInstalled() {
			options = append(options, tui.MenuOption{Label: "Status", Value: actions.ActionServerStatus})
			options = append(options, tui.MenuOption{Label: "Uninstall", Value: actions.ActionServerUninstall})
		}
		options = append(options,
			tui.MenuOption{Label: "", Separator: true},
			tui.MenuOption{Label: "Exit", Value: "exit"},
		)

		choice, err := tui.RunMenu(tui.MenuConfig{
			Title:   "Nethopper Server",
			Options: options,
		})
		if err != nil || choice == "" || choice == "exit" {
			return nil
		}

		runMenuAction(choice)
	}
}

func runClientMenu() error {
	for {
		state := getClientState()
		var options []tui.MenuOption

		if !state.hasBinary {
			options = append(options, tui.MenuOption{Label: "Install", Value: actions.ActionClientInstall})
		} else {
			options = append(options, tui.MenuOption{Label: "Configure", Value: actions.ActionClientConfigure})
		}
		if state.hasConfig {
			options = append(options, tui.MenuOption{Label: "Run", Value: actions.ActionClientRun})
		}
		if state.isInstalled() {
			options = append(options, tui.MenuOption{Label: "Status", Value: actions.ActionClientStatus})
			options = append(options, tui.MenuOption{Label: "Uninstall", Value: actions.ActionClientUninstall})
		}
		options = append(options,
			tui.MenuOption{Label: "", Separator: true},
			tui.MenuOption{Label: "Exit", Value: "exit"},
		)

		choice, err := tui.RunMenu(tui.MenuConfig{
			Title:   "Nethopper Client",
			Options: options,
		})
		if err != nil || choice == "" || choice == "exit" {
			return nil
		}

		runMenuAction(choice)
	}
}

// runMenuAction handles running an action from the menu.
// Inline actions temporarily exit the alt screen to show output directly.
func runMenuAction(actionID string) {
	if isInlineAction(actionID) {
		tui.EndSession()
		err := RunAction(actionID, true)
		if err != nil && !errors.Is(err, errCancelled) {
			tui.PrintError(err.Error())
		}
		tui.WaitForEnter()
		tui.BeginSession()
		return
	}

	if err := RunAction(actionID); err != nil {
		if errors.Is(err, errCancelled) {
			return
		}
		_ = tui.ShowMessage(tui.AppMessage{Type: "error", Message: err.Error()})
	}
}
