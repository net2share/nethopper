//go:build linux

package binary

import (
	"fmt"
	"os/exec"
	"strings"
)

// SetNetBindCapability grants CAP_NET_BIND_SERVICE capability to the binary,
// allowing it to bind to privileged ports (< 1024) without running as root.
// This is a security best practice used by nginx, Apache, BIND, and other
// network services that need to bind to low ports.
//
// The capability is set on the file itself (file capability), which works
// even with systemd's NoNewPrivileges=true security hardening.
//
// Requires: libcap2-bin package (provides setcap command)
// Available on: All Linux kernels since 2.6.24 (2008)
func SetNetBindCapability(binaryPath string) error {
	// Check if setcap is available
	setcapPath, err := exec.LookPath("setcap")
	if err != nil {
		return fmt.Errorf("setcap not found (install libcap2-bin): %w", err)
	}

	// Set CAP_NET_BIND_SERVICE capability
	// +ep = effective and permitted capability sets
	cmd := exec.Command(setcapPath, "cap_net_bind_service=+ep", binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set capability: %s: %w", string(output), err)
	}

	return nil
}

// HasNetBindCapability checks if the binary has CAP_NET_BIND_SERVICE capability set
func HasNetBindCapability(binaryPath string) (bool, error) {
	getcapPath, err := exec.LookPath("getcap")
	if err != nil {
		return false, fmt.Errorf("getcap not found (install libcap2-bin): %w", err)
	}

	cmd := exec.Command(getcapPath, binaryPath)
	output, err := cmd.Output()
	if err != nil {
		// getcap returns error if no capabilities are set
		return false, nil
	}

	// Output looks like: "/path/to/binary cap_net_bind_service=ep"
	return strings.Contains(string(output), "cap_net_bind_service"), nil
}
