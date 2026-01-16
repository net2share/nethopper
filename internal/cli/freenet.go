package cli

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/nethopper/nethopper/internal/commands/freenet"
	"github.com/nethopper/nethopper/internal/config"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/spf13/cobra"
)

var freenetCmd = &cobra.Command{
	Use:   "freenet",
	Short: "Free internet node mode (Windows/macOS/Linux)",
	Long: `Freenet mode runs sing-box on a device with two network interfaces:
one connected to free internet, and one connected to the restricted network.

Traffic from the server is tunneled through this node to access the internet.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !nonInteractive {
			runFreenetMenu()
		} else {
			cmd.Help()
		}
	},
}

// Freenet command flags
var (
	connectionString    string
	serverIP            string
	freenetServerPort   uint16
	freenetVmessPort    uint16
	vmessUUID           string
	freeInterface       string
	restrictedInterface string
	maxConnections      int
	freenetLogLevel     string
	freenetForce        bool
	freenetKeepConfig   bool
	freenetJSON         bool
	freenetPing         bool
)

func init() {
	// Run command
	freenetRunCmd := &cobra.Command{
		Use:   "run",
		Short: "Run sing-box in foreground",
		Long: `Run starts sing-box with the freenet configuration.
If no configuration exists, it will prompt for setup first.

Press Ctrl+C to stop.`,
		Run: runFreenetRun,
	}

	// Configure command
	freenetConfigureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Set up freenet node configuration",
		Long: `Configure sets up the freenet node connection to the server.

You can provide a connection string from the server or enter values manually.
All server-related values (server-port, vmess-port, vmess-uuid) are required.`,
		Run: runFreenetConfigure,
	}
	freenetConfigureCmd.Flags().StringVar(&connectionString, "connection", "", "Connection string from server (nh://...)")
	freenetConfigureCmd.Flags().StringVar(&serverIP, "server-ip", "", "Server IP address")
	freenetConfigureCmd.Flags().Uint16Var(&freenetServerPort, "server-port", 0, "Server reverse endpoint port (required)")
	freenetConfigureCmd.Flags().Uint16Var(&freenetVmessPort, "vmess-port", 0, "Server VMess port (required)")
	freenetConfigureCmd.Flags().StringVar(&vmessUUID, "vmess-uuid", "", "VMess UUID (required)")
	freenetConfigureCmd.Flags().StringVar(&freeInterface, "free-iface", "", "Interface with free internet access")
	freenetConfigureCmd.Flags().StringVar(&restrictedInterface, "restricted-iface", "", "Interface on restricted network")
	freenetConfigureCmd.Flags().IntVar(&maxConnections, "max-connections", 30, "Maximum multiplex connections")
	freenetConfigureCmd.Flags().StringVar(&freenetLogLevel, "log-level", "warn", "Log level")

	// Uninstall command
	freenetUninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove freenet configuration and binary",
		Long:  `Uninstall removes the sing-box binary and configuration from user space.`,
		Run:   runFreenetUninstall,
	}
	freenetUninstallCmd.Flags().BoolVar(&freenetForce, "force", false, "Skip confirmation prompts")
	freenetUninstallCmd.Flags().BoolVar(&freenetKeepConfig, "keep-config", false, "Keep configuration file")

	// Status command
	freenetStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show freenet node status",
		Long: `Status displays the freenet node configuration, interface status,
and connectivity to the server.`,
		Run: runFreenetStatus,
	}
	freenetStatusCmd.Flags().BoolVar(&freenetJSON, "json", false, "Output in JSON format")
	freenetStatusCmd.Flags().BoolVar(&freenetPing, "ping", false, "Test connectivity to server")

	freenetCmd.AddCommand(freenetRunCmd)
	freenetCmd.AddCommand(freenetConfigureCmd)
	freenetCmd.AddCommand(freenetUninstallCmd)
	freenetCmd.AddCommand(freenetStatusCmd)
}

func runFreenetRun(cmd *cobra.Command, args []string) {
	opts := &freenet.RunOptions{
		Verbose: verbose,
	}

	if err := freenet.Run(opts, Info); err != nil {
		Fatal("Run failed: %v", err)
	}
}

func runFreenetConfigure(cmd *cobra.Command, args []string) {
	opts := &freenet.ConfigureOptions{
		ConnectionString:    connectionString,
		ServerIP:            serverIP,
		ServerPort:          freenetServerPort,
		VMESSPort:           freenetVmessPort,
		VMESSUuid:           vmessUUID,
		FreeInterface:       freeInterface,
		RestrictedInterface: restrictedInterface,
		MaxConnections:      maxConnections,
		LogLevel:            freenetLogLevel,
		NonInteractive:      nonInteractive,
		Verbose:             verbose,
	}

	_, err := freenet.Configure(opts, Info)
	if err != nil {
		Fatal("Configure failed: %v", err)
	}

	Success("Configuration saved!")
}

func runFreenetUninstall(cmd *cobra.Command, args []string) {
	// Confirm unless --force
	if !freenetForce && !nonInteractive {
		confirm, err := PromptConfirm("Are you sure you want to uninstall", false)
		if err != nil || !confirm {
			Info("Uninstall cancelled")
			return
		}
	}

	opts := &freenet.UninstallOptions{
		Force:      freenetForce,
		KeepConfig: freenetKeepConfig,
		Verbose:    verbose,
	}

	if err := freenet.Uninstall(opts, Info); err != nil {
		Fatal("Uninstall failed: %v", err)
	}

	Success("Uninstall complete!")
}

func runFreenetStatus(cmd *cobra.Command, args []string) {
	opts := &freenet.StatusOptions{
		JSON: freenetJSON,
		Ping: freenetPing,
	}

	result, err := freenet.Status(opts)
	if err != nil {
		Fatal("Failed to get status: %v", err)
	}

	if freenetJSON {
		freenet.PrintStatusJSON(result)
	} else {
		freenet.PrintStatus(result, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		}, Green, Yellow, Red)
	}
}

func runFreenetMenu() {
	for {
		prompt := promptui.Select{
			Label: "Freenet Node Mode - Select action",
			Items: []string{
				"Run          - Start sing-box (foreground)",
				"Configure    - Set up connection",
				"Uninstall    - Remove configuration",
				"Status       - Show connection status",
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
			runFreenetRunInteractive()
		case 1:
			runFreenetConfigureInteractive()
		case 2:
			runFreenetUninstall(nil, nil)
		case 3:
			runFreenetStatusInteractive()
		case 4:
			return
		}
	}
}

func runFreenetRunInteractive() {
	// Check if runnable
	if err := freenet.CheckRunnable(); err != nil {
		Error("Cannot run: %v", err)
		Info("Run 'Configure' first to set up the connection")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	opts := &freenet.RunOptions{
		Verbose: verbose,
	}

	if err := freenet.Run(opts, Info); err != nil {
		Error("Run failed: %v", err)
	}

	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func runFreenetConfigureInteractive() {
	// Check available interfaces first
	interfaces, err := freenet.ListAvailableInterfaces()
	if err != nil {
		Error("Failed to list interfaces: %v", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	upCount := platform.CountUpInterfaces(interfaces)
	if upCount < 2 {
		Error("At least 2 network interfaces must be up. Found %d.", upCount)
		Info("Ensure both your free internet and restricted network interfaces are connected.")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	// Load existing configuration to use as defaults
	paths := platform.NewPathProvider()
	configMgr := config.NewManager(paths)

	var existingCfg *config.FreenetConfig
	hasExisting := configMgr.ConfigExists(false)
	if hasExisting {
		existingCfg, err = configMgr.LoadFreenetConfig()
		if err != nil {
			Warn("Could not load existing config: %v", err)
			defaults := config.FreenetDefaults()
			existingCfg = &defaults
			hasExisting = false
		}
	} else {
		defaults := config.FreenetDefaults()
		existingCfg = &defaults
	}

	// Ask for custom config
	useCustom, customPath, err := PromptCustomConfig()
	if err != nil && err != promptui.ErrAbort {
		Error("Error: %v", err)
		return
	}

	if useCustom && customPath != "" {
		opts := &freenet.ConfigureOptions{
			CustomConfigPath: customPath,
			Verbose:          verbose,
		}
		_, err := freenet.Configure(opts, Info)
		if err != nil {
			Error("Configure failed: %v", err)
		} else {
			Success("Configuration saved!")
		}
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	opts := &freenet.ConfigureOptions{
		Verbose: verbose,
	}

	// Show existing config info if available
	if hasExisting && existingCfg.ServerIP != "" {
		Info("Existing configuration: Server %s:%d", existingCfg.ServerIP, existingCfg.ServerPort)
	}

	// Ask for connection string or manual entry
	options := []string{
		"Enter connection string",
		"Configure manually",
	}
	idx, _, err := PromptSelect("Configuration method", options)
	if IsInterrupt(err) {
		return
	}
	if err != nil {
		Error("Error: %v", err)
		return
	}

	if idx == 0 {
		// Connection string - use simple prompt to avoid visual artifacts when pasting
		connStr, err := promptSimple("Connection string (nh://...)", true)
		if err != nil {
			Error("Error: %v", err)
			return
		}
		opts.ConnectionString = connStr
	} else {
		// Manual configuration - use existing values as defaults
		ip, err := PromptString("Server IP address", existingCfg.ServerIP, true)
		if IsInterrupt(err) {
			return
		}
		port, err := PromptPort("Server port", existingCfg.ServerPort, true)
		if IsInterrupt(err) {
			return
		}
		vport, err := PromptPort("VMess port", existingCfg.VMESSPort, true)
		if IsInterrupt(err) {
			return
		}
		uuid, err := PromptString("VMess UUID", existingCfg.VMESSUuid, true)
		if IsInterrupt(err) {
			return
		}
		opts.ServerIP = ip
		opts.ServerPort = port
		opts.VMESSPort = vport
		opts.VMESSUuid = uuid
	}

	// Select interfaces with existing as defaults
	fmt.Println("\nAvailable network interfaces:")
	for i, iface := range interfaces {
		fmt.Printf("  %d. %s\n", i+1, iface.String())
	}
	fmt.Println()

	freeIface, err := PromptInterfaceWithDefault("Select FREE internet interface", interfaces, existingCfg.FreeInterface)
	if IsInterrupt(err) {
		return
	}
	if err != nil {
		Error("Error: %v", err)
		return
	}
	opts.FreeInterface = freeIface.Name

	// Filter out selected interface for second selection
	var remainingIfaces []platform.NetworkInterface
	for _, iface := range interfaces {
		if iface.Name != freeIface.Name {
			remainingIfaces = append(remainingIfaces, iface)
		}
	}

	restrictedIface, err := PromptInterfaceWithDefault("Select RESTRICTED network interface", remainingIfaces, existingCfg.RestrictedInterface)
	if IsInterrupt(err) {
		return
	}
	if err != nil {
		Error("Error: %v", err)
		return
	}
	opts.RestrictedInterface = restrictedIface.Name

	// Other options with existing values as defaults
	maxConn, err := PromptMultiplexValue("Max connections", existingCfg.MaxConnections)
	if IsInterrupt(err) {
		return
	}
	opts.MaxConnections = maxConn

	logLevel, err := PromptLogLevel(existingCfg.LogLevel)
	if IsInterrupt(err) {
		return
	}
	opts.LogLevel = logLevel

	_, err = freenet.Configure(opts, Info)
	if err != nil {
		Error("Configure failed: %v", err)
	} else {
		Success("Configuration saved!")
	}

	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}

func runFreenetStatusInteractive() {
	opts := &freenet.StatusOptions{
		JSON: false,
		Ping: true, // Always ping in interactive mode
	}

	result, err := freenet.Status(opts)
	if err != nil {
		Error("Failed to get status: %v", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	freenet.PrintStatus(result, func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	}, Green, Yellow, Red)

	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
}
