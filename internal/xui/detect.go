package xui

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	xuiServiceName = "x-ui"
	xuiBinaryPath  = "/usr/local/x-ui/x-ui"
)

// PanelInfo holds detected x-ui panel configuration.
type PanelInfo struct {
	Port     int
	BasePath string
	CertFile string
	KeyFile  string
	UseHTTPS bool
	BaseURL  string
}

// IsXUIRunning checks if the x-ui systemd service is active and running.
func IsXUIRunning() bool {
	out, err := exec.Command("systemctl", "is-active", xuiServiceName).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}

// ReadPanelInfo reads x-ui panel settings using the x-ui binary directly.
// Runs: /usr/local/x-ui/x-ui setting -show true (non-interactive)
// and:  /usr/local/x-ui/x-ui setting -getCert true (for TLS detection)
func ReadPanelInfo() (*PanelInfo, error) {
	info := &PanelInfo{}

	// Get port and basePath
	out, err := exec.Command(xuiBinaryPath, "setting", "-show", "true").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read x-ui settings: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "port:") {
			fmt.Sscanf(strings.TrimPrefix(line, "port:"), "%d", &info.Port)
		} else if strings.HasPrefix(line, "webBasePath:") {
			info.BasePath = strings.TrimSpace(strings.TrimPrefix(line, "webBasePath:"))
		}
	}

	if info.Port == 0 {
		return nil, fmt.Errorf("could not detect x-ui panel port")
	}

	// Normalize basePath
	if info.BasePath == "" {
		info.BasePath = "/"
	}
	if !strings.HasPrefix(info.BasePath, "/") {
		info.BasePath = "/" + info.BasePath
	}
	if !strings.HasSuffix(info.BasePath, "/") {
		info.BasePath = info.BasePath + "/"
	}

	// Get cert info for TLS detection
	certOut, err := exec.Command(xuiBinaryPath, "setting", "-getCert", "true").Output()
	if err == nil {
		for _, line := range strings.Split(string(certOut), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "cert:") {
				info.CertFile = strings.TrimSpace(strings.TrimPrefix(line, "cert:"))
			} else if strings.HasPrefix(line, "key:") {
				info.KeyFile = strings.TrimSpace(strings.TrimPrefix(line, "key:"))
			}
		}
	}

	info.UseHTTPS = info.CertFile != ""
	scheme := "http"
	if info.UseHTTPS {
		scheme = "https"
	}
	info.BaseURL = fmt.Sprintf("%s://127.0.0.1:%d%s", scheme, info.Port, info.BasePath)

	return info, nil
}
