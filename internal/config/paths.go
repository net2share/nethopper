package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Server paths (Linux only, requires root).
const (
	ServerBinDir    = "/usr/local/bin"
	ServerConfigDir = "/etc/nethopper"
	ServerStateDir  = "/var/lib/nethopper"
)

func ServerConfigPath() string {
	return filepath.Join(ServerConfigDir, "server.json")
}

func ServerXrayConfigPath() string {
	return filepath.Join(ServerConfigDir, "xray.json")
}


// Client paths (cross-platform, user-level).

func ClientConfigDir() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "nethopper")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "nethopper")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "nethopper")
	}
}

func ClientBinDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "nethopper", "bin")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "bin")
	}
}

func ClientConfigPath() string {
	return filepath.Join(ClientConfigDir(), "client.json")
}

func ClientXrayConfigPath() string {
	return filepath.Join(ClientConfigDir(), "xray.json")
}
