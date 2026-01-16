package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/nethopper/nethopper/internal/commands/server"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/util"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server mode (Linux only, requires root)",
	Long: `Server mode runs sing-box as a systemd service on Linux.

The server accepts MTProto connections from Telegram clients and tunnels
traffic through a reverse connection to the freenet node.

Requires root privileges for systemd service management.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if running on Linux
		if runtime.GOOS != "linux" {
			Fatal("Server mode is only supported on Linux (current: %s)", runtime.GOOS)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !nonInteractive {
			runServerMenu()
		} else {
			cmd.Help()
		}
	},
}

// Server command flags
var (
	serverPort          uint16
	mtprotoPort         uint16
	vmessPort           uint16
	mtprotoSecret       string
	fallbackHost        string
	multiplexPerConn    int
	serverLogLevel      string
	serverForce         bool
	serverKeepConfig    bool
	serverJSON          bool
	serverShowLogs      bool
)

func init() {
	// Install command
	serverInstallCmd := &cobra.Command{
		Use:   "install",
		Short: "Install sing-box as systemd service",
		Long: `Install extracts the sing-box binary, generates configuration,
creates a systemd service, and starts it.

Ports are auto-selected if not provided (from range 1024-49151).`,
		Run: runServerInstall,
	}
	serverInstallCmd.Flags().Uint16Var(&serverPort, "port", 0, "Reverse endpoint listen port (auto-select if 0)")
	serverInstallCmd.Flags().Uint16Var(&mtprotoPort, "mtproto-port", 0, "MTProto inbound port (auto-select if 0)")
	serverInstallCmd.Flags().Uint16Var(&vmessPort, "vmess-port", 0, "Outbound VMess port (auto-select if 0)")
	serverInstallCmd.Flags().StringVar(&mtprotoSecret, "secret", "", "MTProto secret (32 hex chars, auto-generate if empty)")
	serverInstallCmd.Flags().StringVar(&fallbackHost, "fallback", "storage.googleapis.com", "Fallback host for TLS camouflage")
	serverInstallCmd.Flags().IntVar(&multiplexPerConn, "multiplex", 50, "Streams per connection")
	serverInstallCmd.Flags().StringVar(&serverLogLevel, "log-level", "warn", "Log level (trace/debug/info/warn/error)")

	// Configure command
	serverConfigureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Modify server configuration",
		Long:  `Configure allows changing server settings. Restarts the service if running.`,
		Run:   runServerConfigure,
	}
	serverConfigureCmd.Flags().Uint16Var(&serverPort, "port", 0, "Reverse endpoint listen port")
	serverConfigureCmd.Flags().Uint16Var(&mtprotoPort, "mtproto-port", 0, "MTProto inbound port")
	serverConfigureCmd.Flags().Uint16Var(&vmessPort, "vmess-port", 0, "Outbound VMess port")
	serverConfigureCmd.Flags().StringVar(&mtprotoSecret, "secret", "", "MTProto secret (32 hex chars)")
	serverConfigureCmd.Flags().StringVar(&fallbackHost, "fallback", "", "Fallback host for TLS camouflage")
	serverConfigureCmd.Flags().IntVar(&multiplexPerConn, "multiplex", 0, "Streams per connection")
	serverConfigureCmd.Flags().StringVar(&serverLogLevel, "log-level", "", "Log level")

	// Uninstall command
	serverUninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove sing-box service and files",
		Long:  `Uninstall stops the service, removes the systemd unit, binary, and configuration.`,
		Run:   runServerUninstall,
	}
	serverUninstallCmd.Flags().BoolVar(&serverForce, "force", false, "Skip confirmation prompts")
	serverUninstallCmd.Flags().BoolVar(&serverKeepConfig, "keep-config", false, "Keep configuration file")

	// Status command
	serverStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show server status and configuration",
		Long:  `Status displays the current server state, configuration, and connection string.`,
		Run:   runServerStatus,
	}
	serverStatusCmd.Flags().BoolVar(&serverJSON, "json", false, "Output in JSON format")
	serverStatusCmd.Flags().BoolVar(&serverShowLogs, "logs", false, "Show service logs (journalctl)")

	// Connection string command
	serverConnectionCmd := &cobra.Command{
		Use:   "connection-string",
		Short: "Generate connection string for freenet nodes",
		Run:   runServerConnectionString,
	}

	serverCmd.AddCommand(serverInstallCmd)
	serverCmd.AddCommand(serverConfigureCmd)
	serverCmd.AddCommand(serverUninstallCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverConnectionCmd)
}

func runServerInstall(cmd *cobra.Command, args []string) {
	if err := server.EnsureRoot(); err != nil {
		Fatal("%v", err)
	}

	opts := &server.InstallOptions{
		ListenPort:             serverPort,
		MTProtoPort:            mtprotoPort,
		VMESSPort:              vmessPort,
		MTProtoSecret:          mtprotoSecret,
		FallbackHost:           fallbackHost,
		MultiplexPerConnection: multiplexPerConn,
		LogLevel:               serverLogLevel,
		NonInteractive:         nonInteractive,
		Verbose:                verbose,
	}

	result, err := server.Install(opts, Info)
	if err != nil {
		Fatal("Install failed: %v", err)
	}

	Success("Installation complete!")
	fmt.Println()
	fmt.Printf("Connection string for freenet nodes:\n  %s\n", Green(result.ConnectionString))
	fmt.Println()
}

func runServerConfigure(cmd *cobra.Command, args []string) {
	if err := server.EnsureRoot(); err != nil {
		Fatal("%v", err)
	}

	opts := &server.ConfigureOptions{
		ListenPort:             serverPort,
		MTProtoPort:            mtprotoPort,
		VMESSPort:              vmessPort,
		MTProtoSecret:          mtprotoSecret,
		FallbackHost:           fallbackHost,
		MultiplexPerConnection: multiplexPerConn,
		LogLevel:               serverLogLevel,
		NonInteractive:         nonInteractive,
		Verbose:                verbose,
	}

	_, err := server.Configure(opts, Info)
	if err != nil {
		Fatal("Configure failed: %v", err)
	}

	Success("Configuration updated!")
}

func runServerUninstall(cmd *cobra.Command, args []string) {
	if err := server.EnsureRoot(); err != nil {
		Fatal("%v", err)
	}

	// Confirm unless --force
	if !serverForce && !nonInteractive {
		confirm, err := PromptConfirm("Are you sure you want to uninstall", false)
		if err != nil || !confirm {
			Info("Uninstall cancelled")
			return
		}
	}

	opts := &server.UninstallOptions{
		Force:      serverForce,
		KeepConfig: serverKeepConfig,
		Verbose:    verbose,
	}

	if err := server.Uninstall(opts, Info); err != nil {
		Fatal("Uninstall failed: %v", err)
	}

	Success("Uninstall complete!")
}

func runServerStatus(cmd *cobra.Command, args []string) {
	// Show logs if requested
	if serverShowLogs {
		if err := server.ViewLogs(true, 100); err != nil {
			Fatal("Failed to view logs: %v", err)
		}
		return
	}

	opts := &server.StatusOptions{
		JSON: serverJSON,
	}

	result, err := server.Status(opts)
	if err != nil {
		Fatal("Failed to get status: %v", err)
	}

	if serverJSON {
		server.PrintStatusJSON(result)
	} else {
		server.PrintStatus(result, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		}, Green, Yellow, Red)
	}
}

func runServerConnectionString(cmd *cobra.Command, args []string) {
	connStr, err := server.GetConnectionString("")
	if err != nil {
		Fatal("Failed to get connection string: %v", err)
	}
	fmt.Println(connStr)
}

func runServerMenu() {
	if err := server.EnsureRoot(); err != nil {
		Error("%v", err)
		return
	}

	for {
		prompt := promptui.Select{
			Label: "Server Mode - Select action",
			Items: []string{
				"Install      - Install sing-box as systemd service",
				"Configure    - Modify server configuration",
				"Uninstall    - Remove sing-box service",
				"Status       - Show service status",
				"Back",
			},
			Templates: selectTemplates,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return
			}
			Fatal("Menu error: %v", err)
		}

		switch idx {
		case 0:
			runServerInstallInteractive()
		case 1:
			runServerConfigureInteractive()
		case 2:
			runServerUninstall(nil, nil)
		case 3:
			runServerStatus(nil, nil)
		case 4:
			return
		}
	}
}

func runServerInstallInteractive() {
	// Check if already installed
	installed, hasConfig, _ := server.CheckInstallation()
	if installed {
		Error("nethopper is already installed. Use 'Configure' to modify settings or 'Uninstall' first.")
		return
	}

	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)

	var port, mtproto, vmess uint16
	var err error

	// Check if config already exists (from running configure first)
	if hasConfig {
		existingCfg, loadErr := configMgr.LoadServerConfig()
		if loadErr == nil && existingCfg.ListenPort > 0 {
			Info("Using existing configuration:")
			Info("  Listen port: %d", existingCfg.ListenPort)
			Info("  MTProto port: %d", existingCfg.MTProtoPort)
			Info("  Outbound VMess port: %d", existingCfg.VMESSPort)
			port = existingCfg.ListenPort
			mtproto = existingCfg.MTProtoPort
			vmess = existingCfg.VMESSPort
		}
	}

	// If no existing config, prompt for ports
	if port == 0 {
		defaults := config.ServerDefaults()

		// Pre-calculate default ports for consistent display
		// This ensures the prompt shows actual port values and displays them
		// consistently when user accepts defaults
		defaultPorts, err := util.FindMultipleAvailablePorts(3, 15)
		if err != nil || len(defaultPorts) < 3 {
			Error("Failed to find available ports: %v", err)
			fmt.Println("Press Enter to continue...")
			fmt.Scanln()
			return
		}

		port, err = PromptPort("Listen port", defaultPorts[0], false)
		if IsInterrupt(err) {
			return
		}
		mtproto, err = PromptPort("MTProto port", defaultPorts[1], false)
		if IsInterrupt(err) {
			return
		}
		vmess, err = PromptPort("Outbound VMess port", defaultPorts[2], false)
		if IsInterrupt(err) {
			return
		}

		// Use defaults for other settings (can be changed via configure)
		Info("Using defaults: Fallback=%s, Multiplex=%d, LogLevel=%s",
			defaults.FallbackHost, defaults.MultiplexPerConnection, defaults.LogLevel)
	}

	opts := &server.InstallOptions{
		ListenPort:     port,
		MTProtoPort:    mtproto,
		VMESSPort:      vmess,
		NonInteractive: false,
		Verbose:        verbose,
	}

	result, err := server.Install(opts, Info)
	if err != nil {
		Error("Install failed: %v", err)
		return
	}

	Success("Installation complete!")
	fmt.Println()
	fmt.Printf("Connection string for freenet nodes:\n  %s\n", Green(result.ConnectionString))
	fmt.Println()
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func runServerConfigureInteractive() {
	// Ask for custom config first
	useCustom, customPath, err := PromptCustomConfig()
	if IsInterrupt(err) {
		return
	}

	if useCustom && customPath != "" {
		// Read the file content
		customContent, err := os.ReadFile(customPath)
		if err != nil {
			Error("Failed to read config file: %v", err)
			fmt.Println("Press Enter to continue...")
			fmt.Scanln()
			return
		}

		opts := &server.ConfigureOptions{
			CustomConfig: customContent,
			Verbose:      verbose,
		}
		_, err = server.Configure(opts, Info)
		if err != nil {
			Error("Configure failed: %v", err)
		} else {
			Success("Configuration saved!")
		}
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	// Load existing configuration to use as defaults
	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)

	var existingCfg *config.ServerConfig
	if configMgr.ConfigExists(true) {
		existingCfg, err = configMgr.LoadServerConfig()
		if err != nil {
			Warn("Could not load existing config, using defaults: %v", err)
			defaults := config.ServerDefaults()
			existingCfg = &defaults
		}
	} else {
		defaults := config.ServerDefaults()
		existingCfg = &defaults
	}

	// If ports are not configured, pre-calculate defaults for consistent display
	if existingCfg.ListenPort == 0 || existingCfg.MTProtoPort == 0 || existingCfg.VMESSPort == 0 {
		defaultPorts, portErr := util.FindMultipleAvailablePorts(3, 15)
		if portErr == nil && len(defaultPorts) >= 3 {
			if existingCfg.ListenPort == 0 {
				existingCfg.ListenPort = defaultPorts[0]
			}
			if existingCfg.MTProtoPort == 0 {
				existingCfg.MTProtoPort = defaultPorts[1]
			}
			if existingCfg.VMESSPort == 0 {
				existingCfg.VMESSPort = defaultPorts[2]
			}
		}
	}

	// Prompt with existing values as defaults
	port, err := PromptPort("Listen port", existingCfg.ListenPort, false)
	if IsInterrupt(err) {
		return
	}
	mtproto, err := PromptPort("MTProto port", existingCfg.MTProtoPort, false)
	if IsInterrupt(err) {
		return
	}
	vmess, err := PromptPort("Outbound VMess port", existingCfg.VMESSPort, false)
	if IsInterrupt(err) {
		return
	}

	// Ask about MTProto secret - keep existing or regenerate
	var regenerateSecret bool
	if existingCfg.MTProtoSecret != "" {
		regenerateSecret, err = PromptConfirm("Regenerate MTProto secret", false)
		if IsInterrupt(err) {
			return
		}
	}

	fallback, err := PromptString("Fallback host", existingCfg.FallbackHost, false)
	if IsInterrupt(err) {
		return
	}
	multiplex, err := PromptMultiplexValue("Streams per connection", existingCfg.MultiplexPerConnection)
	if IsInterrupt(err) {
		return
	}
	logLevel, err := PromptLogLevel(existingCfg.LogLevel)
	if IsInterrupt(err) {
		return
	}

	opts := &server.ConfigureOptions{
		ListenPort:              port,
		MTProtoPort:             mtproto,
		VMESSPort:               vmess,
		RegenerateMTProtoSecret: regenerateSecret,
		FallbackHost:            fallback,
		MultiplexPerConnection:  multiplex,
		LogLevel:                logLevel,
		Verbose:                 verbose,
	}

	_, err = server.Configure(opts, Info)
	if err != nil {
		Error("Configure failed: %v", err)
		return
	}

	Success("Configuration updated!")
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}
