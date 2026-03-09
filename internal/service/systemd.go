package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	ServiceName  = "nethopper"
	ServiceUser  = "nethopper"
	ServiceGroup = "nethopper"
)

const systemdTemplate = `[Unit]
Description=nethopper xray proxy service
After=network.target nss-lookup.target
Wants=network-online.target

[Service]
Type=simple
User={{.User}}
Group={{.Group}}
ExecStart={{.BinaryPath}} run -c {{.ConfigPath}}
Restart=on-failure
RestartSec=10
LimitNOFILE=infinity
WorkingDirectory={{.WorkingDir}}

ProtectSystem=strict
ProtectHome=true
ReadWritePaths={{.StateDir}}
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`

// ServiceConfig contains configuration for creating a systemd service.
type ServiceConfig struct {
	BinaryPath string
	ConfigPath string
	User       string
	Group      string
	WorkingDir string
	StateDir   string
}

// ServiceStatus represents the status of a systemd service.
type ServiceStatus struct {
	Running   bool
	Enabled   bool
	Active    string
	LoadState string
	SubState  string
}

// SystemdManager manages systemd services.
type SystemdManager struct {
	unitPath string
}

func NewSystemdManager() *SystemdManager {
	return &SystemdManager{
		unitPath: filepath.Join("/etc/systemd/system", ServiceName+".service"),
	}
}

func IsSystemd() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

func (m *SystemdManager) Install(config *ServiceConfig) error {
	if config.User == "" {
		config.User = ServiceUser
	}
	if config.Group == "" {
		config.Group = ServiceGroup
	}
	if config.WorkingDir == "" {
		config.WorkingDir = filepath.Dir(config.ConfigPath)
	}
	if config.StateDir == "" {
		config.StateDir = "/var/lib/nethopper"
	}

	if err := m.ensureServiceUser(config.User, config.Group); err != nil {
		return fmt.Errorf("failed to create service user: %w", err)
	}

	if err := os.MkdirAll(config.StateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	m.chown(config.StateDir, config.User, config.Group)

	tmpl, err := template.New("service").Parse(systemdTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	if err := os.WriteFile(m.unitPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return m.daemonReload()
}

func (m *SystemdManager) Uninstall() error {
	_ = m.Stop()
	_ = m.Disable()

	if err := os.Remove(m.unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	return m.daemonReload()
}

func (m *SystemdManager) Start() error   { return m.systemctl("start", ServiceName) }
func (m *SystemdManager) Stop() error    { return m.systemctl("stop", ServiceName) }
func (m *SystemdManager) Restart() error { return m.systemctl("restart", ServiceName) }
func (m *SystemdManager) Enable() error  { return m.systemctl("enable", ServiceName) }
func (m *SystemdManager) Disable() error { return m.systemctl("disable", ServiceName) }

func (m *SystemdManager) Status() (*ServiceStatus, error) {
	status := &ServiceStatus{}

	output, err := exec.Command("systemctl", "is-active", ServiceName).Output()
	if err == nil {
		status.Active = strings.TrimSpace(string(output))
		status.Running = status.Active == "active"
	} else {
		status.Active = "inactive"
	}

	output, _ = exec.Command("systemctl", "is-enabled", ServiceName).Output()
	status.Enabled = strings.TrimSpace(string(output)) == "enabled"

	output, _ = exec.Command("systemctl", "show", "-p", "LoadState", "--value", ServiceName).Output()
	status.LoadState = strings.TrimSpace(string(output))

	output, _ = exec.Command("systemctl", "show", "-p", "SubState", "--value", ServiceName).Output()
	status.SubState = strings.TrimSpace(string(output))

	return status, nil
}

func (m *SystemdManager) IsInstalled() bool {
	_, err := os.Stat(m.unitPath)
	return err == nil
}

func (m *SystemdManager) systemctl(command, service string) error {
	cmd := exec.Command("systemctl", command, service)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s %s failed: %s", command, service, string(output))
	}
	return nil
}

func (m *SystemdManager) daemonReload() error {
	return exec.Command("systemctl", "daemon-reload").Run()
}

func (m *SystemdManager) ensureServiceUser(username, groupname string) error {
	_, err := user.Lookup(username)
	if err == nil {
		return nil
	}

	exec.Command("groupadd", "-r", groupname).Run()

	cmd := exec.Command("useradd", "-r", "-g", groupname, "-s", "/sbin/nologin",
		"-d", "/var/lib/nethopper", "-c", "nethopper service user", username)
	if err := cmd.Run(); err != nil {
		if _, lookupErr := user.Lookup(username); lookupErr != nil {
			return fmt.Errorf("failed to create user %s: %w", username, err)
		}
	}
	return nil
}

func (m *SystemdManager) chown(path, username, groupname string) {
	exec.Command("chown", "-R", fmt.Sprintf("%s:%s", username, groupname), path).Run()
}
