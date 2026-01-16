//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

// WindowsPathProvider implements PathProvider for Windows
type WindowsPathProvider struct{}

// NewPathProvider creates a new PathProvider for the current platform
func NewPathProvider() PathProvider {
	return &WindowsPathProvider{}
}

func getAppData() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return appData
	}
	return filepath.Join(GetHomeDir(), "AppData", "Roaming")
}

func getProgramData() string {
	if programData := os.Getenv("PROGRAMDATA"); programData != "" {
		return programData
	}
	return "C:\\ProgramData"
}

func (p *WindowsPathProvider) SystemBinaryDir() string {
	// For system-wide installation (requires admin)
	return filepath.Join(getProgramData(), "nethopper", "bin")
}

func (p *WindowsPathProvider) UserBinaryDir() string {
	return filepath.Join(getAppData(), "nethopper", "bin")
}

func (p *WindowsPathProvider) SystemConfigDir() string {
	return filepath.Join(getProgramData(), "nethopper")
}

func (p *WindowsPathProvider) UserConfigDir() string {
	return filepath.Join(getAppData(), "nethopper")
}

func (p *WindowsPathProvider) SystemStateDir() string {
	return filepath.Join(getProgramData(), "nethopper", "state")
}

func (p *WindowsPathProvider) UserStateDir() string {
	return filepath.Join(getAppData(), "nethopper", "state")
}

func (p *WindowsPathProvider) BinaryPath(systemMode bool) string {
	dir := p.UserBinaryDir()
	if systemMode {
		dir = p.SystemBinaryDir()
	}
	return filepath.Join(dir, BinaryName+".exe")
}

func (p *WindowsPathProvider) ConfigPath(systemMode bool) string {
	dir := p.UserConfigDir()
	configFile := FreenetConfigFile
	if systemMode {
		dir = p.SystemConfigDir()
		configFile = ServerConfigFile
	}
	return filepath.Join(dir, configFile)
}

func (p *WindowsPathProvider) StatePath(systemMode bool) string {
	dir := p.UserStateDir()
	if systemMode {
		dir = p.SystemStateDir()
	}
	return filepath.Join(dir, StateFile)
}
