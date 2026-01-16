//go:build linux

package platform

import (
	"path/filepath"
)

// LinuxPathProvider implements PathProvider for Linux
type LinuxPathProvider struct{}

// NewPathProvider creates a new PathProvider for the current platform
func NewPathProvider() PathProvider {
	return &LinuxPathProvider{}
}

func (p *LinuxPathProvider) SystemBinaryDir() string {
	return "/usr/local/bin"
}

func (p *LinuxPathProvider) UserBinaryDir() string {
	return filepath.Join(GetHomeDir(), ".local", "bin")
}

func (p *LinuxPathProvider) SystemConfigDir() string {
	return "/etc/nethopper"
}

func (p *LinuxPathProvider) UserConfigDir() string {
	return filepath.Join(GetConfigHome(), "nethopper")
}

func (p *LinuxPathProvider) SystemStateDir() string {
	return "/var/lib/nethopper"
}

func (p *LinuxPathProvider) UserStateDir() string {
	return filepath.Join(GetStateHome(), "nethopper")
}

func (p *LinuxPathProvider) BinaryPath(systemMode bool) string {
	dir := p.UserBinaryDir()
	if systemMode {
		dir = p.SystemBinaryDir()
	}
	return filepath.Join(dir, BinaryName)
}

func (p *LinuxPathProvider) ConfigPath(systemMode bool) string {
	dir := p.UserConfigDir()
	configFile := FreenetConfigFile
	if systemMode {
		dir = p.SystemConfigDir()
		configFile = ServerConfigFile
	}
	return filepath.Join(dir, configFile)
}

func (p *LinuxPathProvider) StatePath(systemMode bool) string {
	dir := p.UserStateDir()
	if systemMode {
		dir = p.SystemStateDir()
	}
	return filepath.Join(dir, StateFile)
}
