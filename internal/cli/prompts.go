package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/nethopper/nethopper/internal/platform"
	"github.com/nethopper/nethopper/internal/util"
)

// ErrInterrupt is returned when the user presses Ctrl+C during a prompt
var ErrInterrupt = errors.New("interrupted")

// IsInterrupt checks if an error is an interrupt (Ctrl+C)
func IsInterrupt(err error) bool {
	return err == promptui.ErrInterrupt || err == ErrInterrupt
}

// promptTemplates provides consistent styling for all prompts
// Removes default icons to prevent terminal rendering artifacts
var promptTemplates = &promptui.PromptTemplates{
	Prompt:  "{{ . }}: ",
	Valid:   "{{ . }}: ",
	Invalid: "{{ . }}: ",
	Success: "{{ . | faint }}: ",
}

// selectTemplates provides consistent styling for all select prompts
var selectTemplates = &promptui.SelectTemplates{
	Label:    "{{ . }}",
	Active:   "> {{ . }}",
	Inactive: "  {{ . }}",
	Selected: "{{ . | faint }}",
}

// maxDisplayValueLen is the maximum length for values shown in prompt results
const maxDisplayValueLen = 60

// minTerminalWidth is used to estimate line wrapping for clearing prompt lines
const minTerminalWidth = 80

// truncateForDisplay truncates a string for display purposes
func truncateForDisplay(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// clearPromptLines clears the prompt line(s) and displays the result
// It calculates how many terminal lines were used based on prompt+input length
func clearPromptLines(promptLen int, inputLen int, label string, displayValue string) {
	// Calculate total characters displayed (prompt + ": " + input)
	totalLen := promptLen + 2 + inputLen

	// Estimate number of lines (conservative estimate using minimum terminal width)
	lines := (totalLen + minTerminalWidth - 1) / minTerminalWidth
	if lines < 1 {
		lines = 1
	}

	// Move up and clear each line
	for i := 0; i < lines; i++ {
		fmt.Printf("\033[1A\033[K") // Move cursor up and clear line
	}

	// Print the result
	fmt.Printf("\033[2m%s: %s\033[0m\n", label, truncateForDisplay(displayValue, maxDisplayValueLen))
}

// promptSimple provides a basic prompt without interactive redrawing
// Use this for inputs that may be long (like connection strings) to avoid
// visual artifacts when pasting
func promptSimple(label string, required bool) (string, error) {
	fmt.Printf("%s: ", label)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	if required && input == "" {
		return "", fmt.Errorf("value is required")
	}

	// Move up and show grayed result
	fmt.Printf("\033[1A\033[K") // Move cursor up and clear line
	fmt.Printf("\033[2m%s: %s\033[0m\n", label, truncateForDisplay(input, maxDisplayValueLen))

	return input, nil
}

// PromptString prompts for a string value with optional default
// Shows default in brackets: "Label [default]: "
// After Enter, overwrites line to show actual value used
func PromptString(label string, defaultValue string, required bool) (string, error) {
	displayLabel := label
	if defaultValue != "" {
		displayLabel = fmt.Sprintf("%s [%s]", label, defaultValue)
	}

	prompt := promptui.Prompt{
		Label:     displayLabel,
		Templates: promptTemplates,
	}

	if required && defaultValue == "" {
		prompt.Validate = func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("value is required")
			}
			return nil
		}
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// Determine actual value
	actualValue := result
	if result == "" && defaultValue != "" {
		actualValue = defaultValue
	}

	// Clear prompt line(s) and show result
	clearPromptLines(len(displayLabel), len(result), label, actualValue)

	return actualValue, nil
}

// PromptPort prompts for a port number with validation
// Shows default in brackets: "Label [default] (1-65535): "
// After Enter, overwrites line to show actual value used
func PromptPort(label string, defaultValue uint16, required bool) (uint16, error) {
	displayLabel := label
	if defaultValue > 0 {
		displayLabel = fmt.Sprintf("%s [%d] (%s)", label, defaultValue, util.PortRangeString())
	} else {
		displayLabel = fmt.Sprintf("%s (%s)", label, util.PortRangeString())
	}

	prompt := promptui.Prompt{
		Label:     displayLabel,
		Templates: promptTemplates,
		Validate: func(input string) error {
			if input == "" {
				if required && defaultValue == 0 {
					return fmt.Errorf("port is required")
				}
				return nil
			}
			port, err := strconv.Atoi(input)
			if err != nil {
				return fmt.Errorf("invalid port number")
			}
			if port < 1 || port > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	// Determine actual value
	var actualValue uint16
	if result == "" {
		actualValue = defaultValue
	} else {
		port, _ := strconv.Atoi(result)
		actualValue = uint16(port)
	}

	// Clear prompt line(s) and show result
	clearPromptLines(len(displayLabel), len(result), label, fmt.Sprintf("%d", actualValue))

	return actualValue, nil
}

// PromptConfirm prompts for yes/no confirmation
// After Enter, overwrites line to show actual value used
func PromptConfirm(label string, defaultYes bool) (bool, error) {
	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}

	displayLabel := fmt.Sprintf("%s [%s]", label, hint)
	prompt := promptui.Prompt{
		Label:     displayLabel,
		Templates: promptTemplates,
	}

	result, err := prompt.Run()
	if err != nil {
		return defaultYes, err
	}

	// Determine actual value
	var actualValue bool
	inputLen := len(result)
	result = strings.ToLower(strings.TrimSpace(result))
	if result == "" {
		actualValue = defaultYes
	} else {
		actualValue = result == "y" || result == "yes"
	}

	// Clear prompt line(s) and show result
	displayValue := "no"
	if actualValue {
		displayValue = "yes"
	}
	clearPromptLines(len(displayLabel), inputLen, label, displayValue)

	return actualValue, nil
}

// PromptSelect prompts to select from a list of options
func PromptSelect(label string, items []string) (int, string, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     items,
		Templates: selectTemplates,
	}

	return prompt.Run()
}

// PromptInterface prompts to select a network interface
func PromptInterface(label string, interfaces []platform.NetworkInterface) (*platform.NetworkInterface, error) {
	return PromptInterfaceWithDefault(label, interfaces, "")
}

// PromptInterfaceWithDefault prompts to select a network interface with a default selection
// Shows current value in brackets: "Label [current]: "
func PromptInterfaceWithDefault(label string, interfaces []platform.NetworkInterface, defaultName string) (*platform.NetworkInterface, error) {
	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no interfaces available")
	}

	items := make([]string, len(interfaces))
	defaultIdx := 0
	for i, iface := range interfaces {
		items[i] = iface.String()
		if iface.Name == defaultName {
			defaultIdx = i
		}
	}

	displayLabel := label
	if defaultName != "" {
		displayLabel = fmt.Sprintf("%s [%s]", label, defaultName)
	}

	prompt := promptui.Select{
		Label:     displayLabel,
		Items:     items,
		CursorPos: defaultIdx,
		Templates: selectTemplates,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &interfaces[idx], nil
}

// PromptSecret prompts for a secret value with masking option
func PromptSecret(label string, mask bool) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		Templates: promptTemplates,
	}

	if mask {
		prompt.Mask = '*'
	}

	return prompt.Run()
}

// PromptLogLevel prompts for log level selection
// Shows current value in brackets: "Log level [current]: "
func PromptLogLevel(defaultLevel string) (string, error) {
	levels := []string{"trace", "debug", "info", "warn", "error", "disabled"}

	// Find default index
	defaultIdx := 3 // warn
	for i, level := range levels {
		if level == defaultLevel {
			defaultIdx = i
			break
		}
	}

	displayLabel := "Log level"
	if defaultLevel != "" {
		displayLabel = fmt.Sprintf("Log level [%s]", defaultLevel)
	}

	prompt := promptui.Select{
		Label:     displayLabel,
		Items:     levels,
		CursorPos: defaultIdx,
		Templates: selectTemplates,
	}

	_, result, err := prompt.Run()
	return result, err
}

// PromptMultiplexValue prompts for a multiplexing value with default
// Shows default in brackets: "Label [default]: "
// After Enter, overwrites line to show actual value used
func PromptMultiplexValue(label string, defaultValue int) (int, error) {
	displayLabel := fmt.Sprintf("%s [%d]", label, defaultValue)

	prompt := promptui.Prompt{
		Label:     displayLabel,
		Templates: promptTemplates,
		Validate: func(input string) error {
			if input == "" {
				return nil // Will use default
			}
			val, err := strconv.Atoi(input)
			if err != nil {
				return fmt.Errorf("invalid number")
			}
			if val < 1 || val > 1000 {
				return fmt.Errorf("value must be between 1 and 1000")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return defaultValue, err
	}

	// Determine actual value
	var actualValue int
	if result == "" {
		actualValue = defaultValue
	} else {
		actualValue, _ = strconv.Atoi(result)
	}

	// Clear prompt line(s) and show result
	clearPromptLines(len(displayLabel), len(result), label, fmt.Sprintf("%d", actualValue))

	return actualValue, nil
}

// PromptConnectionString prompts for connection string or manual entry
func PromptConnectionString() (string, bool, error) {
	options := []string{
		"Enter connection string from server",
		"Configure manually",
	}

	idx, _, err := PromptSelect("Configuration method", options)
	if err != nil {
		return "", false, err
	}

	if idx == 0 {
		// Use simple prompt to avoid visual artifacts when pasting long strings
		result, err := promptSimple("Connection string (nh://...)", true)
		if err != nil {
			return "", false, err
		}

		// Validate
		if !strings.HasPrefix(result, "nh://") {
			return "", false, fmt.Errorf("connection string must start with 'nh://'")
		}

		return result, true, nil
	}

	return "", false, nil
}

// PromptCustomConfig prompts whether to use custom config file
func PromptCustomConfig() (bool, string, error) {
	confirm, err := PromptConfirm("Do you want to provide a custom config file", false)
	if err != nil || !confirm {
		return false, "", err
	}

	// Ask for file path - use simple prompt for potentially long paths
	path, err := promptSimple("Config file path", true)
	if err != nil {
		return false, "", err
	}

	return true, path, nil
}
