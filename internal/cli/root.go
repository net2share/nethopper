package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	configPath     string
	verbose        bool
	noColor        bool
	nonInteractive bool

	// Version info (set at build time)
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Color outputs
var (
	greenColor  = color.New(color.FgGreen).SprintFunc()
	yellowColor = color.New(color.FgYellow).SprintFunc()
	redColor    = color.New(color.FgRed).SprintFunc()
	cyanColor   = color.New(color.FgCyan).SprintFunc()
	boldColor   = color.New(color.Bold).SprintFunc()
)

// Color wrapper functions that accept string and return string
func Green(s string) string  { return greenColor(s) }
func Yellow(s string) string { return yellowColor(s) }
func Red(s string) string    { return redColor(s) }
func Cyan(s string) string   { return cyanColor(s) }
func Bold(s string) string   { return boldColor(s) }

var rootCmd = &cobra.Command{
	Use:   "nethopper",
	Short: "Share internet from free network to restricted network",
	Long: `nethopper is a sing-box wrapper for sharing internet access.

It supports two modes:
  - server:  Run on a Linux server in the restricted network
  - freenet: Run on a device with access to both free and restricted networks

Use 'nethopper [mode] --help' for more information about a mode.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand, show interactive menu
		if !nonInteractive {
			runInteractiveMenu()
		} else {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Override config path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Disable interactive prompts")

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(freenetCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nethopper %s\n", Version)
		fmt.Printf("  Build time: %s\n", BuildTime)
		fmt.Printf("  Git commit: %s\n", GitCommit)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// IsNonInteractive returns whether non-interactive mode is enabled
func IsNonInteractive() bool {
	return nonInteractive
}

// GetConfigPath returns the config path override
func GetConfigPath() string {
	return configPath
}

// Log prints a message if verbose mode is enabled
func Log(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", Cyan("ℹ"), fmt.Sprintf(format, args...))
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", Green("✓"), fmt.Sprintf(format, args...))
}

// Warn prints a warning message
func Warn(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", Yellow("⚠"), fmt.Sprintf(format, args...))
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red("✗"), fmt.Sprintf(format, args...))
}

// Fatal prints an error message and exits
func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
