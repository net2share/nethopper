package cli

import (
	"fmt"
	"runtime"

	"github.com/manifoldco/promptui"
)

// runInteractiveMenu shows the main menu
func runInteractiveMenu() {
	fmt.Println()
	fmt.Println(Bold("Welcome to nethopper - sing-box wrapper for internet sharing"))
	fmt.Println()

	for {
		prompt := promptui.Select{
			Label:     "Select mode",
			Items:     []string{
				"Server in the restricted network",
				"Free Internet node with access to the restricted network",
				"Exit",
			},
			Templates: selectTemplates,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return
			}
			Fatal("Menu error: %v", err)
		}

		switch idx {
		case 0:
			if runtime.GOOS != "linux" {
				Error("Server mode is only supported on Linux (current: %s)", runtime.GOOS)
				fmt.Println("Press Enter to continue...")
				fmt.Scanln()
				continue
			}
			runServerMenu()
		case 1:
			runFreenetMenu()
		case 2:
			return
		}
	}
}
