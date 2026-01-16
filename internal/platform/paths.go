package platform

import (
	"os"
	"path/filepath"
)

// PathProvider defines the interface for platform-specific paths
type PathProvider interface {
	// Binary paths
	SystemBinaryDir() string // For server mode (system-wide installation)
	UserBinaryDir() string   // For freenet mode (user-space installation)

	// Config paths
	SystemConfigDir() string // For server mode
	UserConfigDir() string   // For freenet mode

	// State paths
	SystemStateDir() string // For server mode
	UserStateDir() string   // For freenet mode

	// Get the appropriate paths based on mode
	BinaryPath(systemMode bool) string
	ConfigPath(systemMode bool) string
	StatePath(systemMode bool) string
}

// Filenames
const (
	BinaryName           = "sing-box"
	ServerConfigFile     = "server.json"
	FreenetConfigFile    = "freenet.json"
	StateFile            = "state.json"
	ServiceName          = "nethopper"
)

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// GetConfigHome returns XDG_CONFIG_HOME or default
func GetConfigHome() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	return filepath.Join(GetHomeDir(), ".config")
}

// GetDataHome returns XDG_DATA_HOME or default
func GetDataHome() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}
	return filepath.Join(GetHomeDir(), ".local", "share")
}

// GetStateHome returns XDG_STATE_HOME or default
func GetStateHome() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg
	}
	return filepath.Join(GetHomeDir(), ".local", "state")
}
