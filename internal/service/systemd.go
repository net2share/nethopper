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
	// ServiceName is the name of the systemd service
	ServiceName = "nethopper"
	// ServiceUser is the user to run the service as
	ServiceUser = "nethopper"
	// ServiceGroup is the group to run the service as
	ServiceGroup = "nethopper"
)

// SystemdServiceTemplate is the template for the systemd service file
const SystemdServiceTemplate = `[Unit]
Description=nethopper sing-box proxy service
Documentation=https://github.com/nethopper/nethopper
After=network.target nss-lookup.target
Wants=network-online.target

[Service]
Type=simple
User={{.User}}
Group={{.Group}}
ExecStart={{.BinaryPath}} run -c {{.ConfigPath}}
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=10
LimitNOFILE=infinity
WorkingDirectory={{.WorkingDir}}

# Security hardening
# Allow binding to privileged ports (< 1024) without running as root
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
ProtectSystem=strict
ProtectHome=true
ReadWritePaths={{.StateDir}}
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`

// ServiceConfig contains configuration for creating a systemd service
type ServiceConfig struct {
	BinaryPath string
	ConfigPath string
	User       string
	Group      string
	WorkingDir string
	StateDir   string
}

// ServiceStatus represents the status of a systemd service
type ServiceStatus struct {
	Running   bool
	Enabled   bool
	Active    string // "active", "inactive", "failed", etc.
	LoadState string // "loaded", "not-found", etc.
	SubState  string // "running", "dead", "failed", etc.
}

// SystemdManager manages systemd services
type SystemdManager struct {
	unitPath string
}

// NewSystemdManager creates a new SystemdManager
func NewSystemdManager() *SystemdManager {
	return &SystemdManager{
		unitPath: filepath.Join("/etc/systemd/system", ServiceName+".service"),
	}
}

// IsSystemd checks if systemd is available
func IsSystemd() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

// Install installs the systemd service
func (m *SystemdManager) Install(config *ServiceConfig) error {
	// Set defaults
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

	// Create service user if it doesn't exist
	if err := m.ensureServiceUser(config.User, config.Group); err != nil {
		return fmt.Errorf("failed to create service user: %w", err)
	}

	// Create state directory
	if err := os.MkdirAll(config.StateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Set ownership of state directory
	if err := m.chown(config.StateDir, config.User, config.Group); err != nil {
		return fmt.Errorf("failed to set state directory ownership: %w", err)
	}

	// Generate service file
	tmpl, err := template.New("service").Parse(SystemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	// Write service file
	if err := os.WriteFile(m.unitPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd daemon
	if err := m.daemonReload(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// Uninstall removes the systemd service
func (m *SystemdManager) Uninstall() error {
	// Stop service if running
	_ = m.Stop()

	// Disable service
	_ = m.Disable()

	// Remove service file
	if err := os.Remove(m.unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd daemon
	if err := m.daemonReload(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// Start starts the service
func (m *SystemdManager) Start() error {
	return m.systemctl("start", ServiceName)
}

// Stop stops the service
func (m *SystemdManager) Stop() error {
	return m.systemctl("stop", ServiceName)
}

// Restart restarts the service
func (m *SystemdManager) Restart() error {
	return m.systemctl("restart", ServiceName)
}

// Enable enables the service to start on boot
func (m *SystemdManager) Enable() error {
	return m.systemctl("enable", ServiceName)
}

// Disable disables the service from starting on boot
func (m *SystemdManager) Disable() error {
	return m.systemctl("disable", ServiceName)
}

// Status returns the service status
func (m *SystemdManager) Status() (*ServiceStatus, error) {
	status := &ServiceStatus{}

	// Get active state
	output, err := exec.Command("systemctl", "is-active", ServiceName).Output()
	if err == nil {
		status.Active = strings.TrimSpace(string(output))
		status.Running = status.Active == "active"
	} else {
		status.Active = "inactive"
	}

	// Get enabled state
	output, err = exec.Command("systemctl", "is-enabled", ServiceName).Output()
	if err == nil {
		state := strings.TrimSpace(string(output))
		status.Enabled = state == "enabled"
	}

	// Get load state
	output, err = exec.Command("systemctl", "show", "-p", "LoadState", "--value", ServiceName).Output()
	if err == nil {
		status.LoadState = strings.TrimSpace(string(output))
	}

	// Get sub state
	output, err = exec.Command("systemctl", "show", "-p", "SubState", "--value", ServiceName).Output()
	if err == nil {
		status.SubState = strings.TrimSpace(string(output))
	}

	return status, nil
}

// IsInstalled checks if the service is installed
func (m *SystemdManager) IsInstalled() bool {
	_, err := os.Stat(m.unitPath)
	return err == nil
}

// GetUnitPath returns the path to the service unit file
func (m *SystemdManager) GetUnitPath() string {
	return m.unitPath
}

// ViewLogs displays the service logs using journalctl
func (m *SystemdManager) ViewLogs(follow bool, lines int) error {
	args := []string{"-u", ServiceName}
	if follow {
		args = append(args, "-f")
	}
	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}

	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// systemctl runs a systemctl command
func (m *SystemdManager) systemctl(command, service string) error {
	cmd := exec.Command("systemctl", command, service)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s %s failed: %s", command, service, string(output))
	}
	return nil
}

// daemonReload runs systemctl daemon-reload
func (m *SystemdManager) daemonReload() error {
	return exec.Command("systemctl", "daemon-reload").Run()
}

// ensureServiceUser creates the service user and group if they don't exist
func (m *SystemdManager) ensureServiceUser(username, groupname string) error {
	// Check if user exists
	_, err := user.Lookup(username)
	if err == nil {
		return nil // User already exists
	}

	// Create group first
	cmd := exec.Command("groupadd", "-r", groupname)
	cmd.Run() // Ignore error if group already exists

	// Create user
	cmd = exec.Command("useradd", "-r", "-g", groupname, "-s", "/sbin/nologin", "-d", "/var/lib/nethopper", "-c", "nethopper service user", username)
	if err := cmd.Run(); err != nil {
		// Check if user was created despite error
		_, lookupErr := user.Lookup(username)
		if lookupErr != nil {
			return fmt.Errorf("failed to create user %s: %w", username, err)
		}
	}

	return nil
}

// chown changes ownership of a file or directory
func (m *SystemdManager) chown(path, username, groupname string) error {
	cmd := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", username, groupname), path)
	return cmd.Run()
}
