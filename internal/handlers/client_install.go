package handlers

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/net2share/nethopper/internal/actions"
	"github.com/net2share/nethopper/internal/binary"
	"github.com/net2share/nethopper/internal/config"
)

func HandleClientInstall(ctx *actions.Context) error {
	beginProgress(ctx, "Installing Xray")

	step, total := 1, 1
	if runtime.GOOS == "linux" {
		total = 2
	}

	ctx.Output.Step(step, total, "Downloading xray binary")
	mgr := binary.NewClientManager(config.ClientBinDir())
	xrayPath, err := mgr.EnsureInstalled(binary.XrayDef, nil)
	if err != nil {
		return failProgress(ctx, fmt.Errorf("failed to install xray: %w", err))
	}
	ctx.Output.Success(fmt.Sprintf("Xray binary: %s", xrayPath))
	step++

	// Set CAP_NET_RAW so xray can bind to specific interfaces without root
	if runtime.GOOS == "linux" {
		ctx.Output.Step(step, total, "Setting network capabilities")
		if err := exec.Command("sudo", "setcap", "cap_net_raw+ep", xrayPath).Run(); err != nil {
			ctx.Output.Warning(fmt.Sprintf("Failed to set capabilities: %v (run may require sudo)", err))
		} else {
			ctx.Output.Success("CAP_NET_RAW set on xray binary")
		}
	}

	endProgress(ctx)
	return nil
}
