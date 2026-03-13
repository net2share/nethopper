package handlers

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/net2share/nethopper/internal/actions"
)

// beginProgress starts a progress view in interactive mode.
func beginProgress(ctx *actions.Context, title string) {
	if ctx.IsInteractive {
		ctx.Output.BeginProgress(title)
	}
}

// endProgress ends a progress view in interactive mode.
func endProgress(ctx *actions.Context) {
	if ctx.IsInteractive {
		ctx.Output.EndProgress()
	}
}

// failProgress shows an error in the progress view and returns the error.
func failProgress(ctx *actions.Context, err error) error {
	if ctx.IsInteractive {
		ctx.Output.Error(fmt.Sprintf("Failed: %v", err))
		ctx.Output.EndProgress()
	}
	return err
}

// detectServerIP tries to find the server's IP address.
func detectServerIP() string {
	if ip := os.Getenv("SERVER_IP"); ip != "" {
		return ip
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "<server-ip>"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			return ipNet.IP.String()
		}
	}
	return "<server-ip>"
}

// copyFile copies a file from src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// randomAvailablePort finds a random available TCP port.
func randomAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}

// isPortAvailable checks if a specific TCP port is available for binding.
func isPortAvailable(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	l.Close()
	return true
}

// writeFile writes data to a file, creating parent directories as needed.
func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0640)
}
