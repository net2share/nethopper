//go:build darwin

package platform

import (
	"path/filepath"
)

// DarwinPathProvider implements PathProvider for macOS
type DarwinPathProvider struct{}

// NewPathProvider creates a new PathProvider for the current platform
func NewPathProvider() PathProvider {
	return &DarwinPathProvider{}
}

func (p *DarwinPathProvider) SystemBinaryDir() string {
	return "/usr/local/bin"
}

func (p *DarwinPathProvider) UserBinaryDir() string {
	return filepath.Join(GetHomeDir(), "Library", "Application Support", "nethopper", "bin")
}

func (p *DarwinPathProvider) SystemConfigDir() string {
	return "/etc/nethopper"
}

func (p *DarwinPathProvider) UserConfigDir() string {
	return filepath.Join(GetHomeDir(), "Library", "Application Support", "nethopper")
}

func (p *DarwinPathProvider) SystemStateDir() string {
	return "/var/lib/nethopper"
}

func (p *DarwinPathProvider) UserStateDir() string {
	return filepath.Join(GetHomeDir(), "Library", "Application Support", "nethopper", "state")
}

func (p *DarwinPathProvider) BinaryPath(systemMode bool) string {
	dir := p.UserBinaryDir()
	if systemMode {
		dir = p.SystemBinaryDir()
	}
	return filepath.Join(dir, BinaryName)
}

func (p *DarwinPathProvider) ConfigPath(systemMode bool) string {
	dir := p.UserConfigDir()
	configFile := FreenetConfigFile
	if systemMode {
		dir = p.SystemConfigDir()
		configFile = ServerConfigFile
	}
	return filepath.Join(dir, configFile)
}

func (p *DarwinPathProvider) StatePath(systemMode bool) string {
	dir := p.UserStateDir()
	if systemMode {
		dir = p.SystemStateDir()
	}
	return filepath.Join(dir, StateFile)
}
