package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// ProcessManager manages sing-box processes for freenet mode
type ProcessManager struct {
	binaryPath string
	configPath string
	cmd        *exec.Cmd
	cancel     context.CancelFunc
}

// NewProcessManager creates a new ProcessManager
func NewProcessManager(binaryPath, configPath string) *ProcessManager {
	return &ProcessManager{
		binaryPath: binaryPath,
		configPath: configPath,
	}
}

// Run starts sing-box in foreground and blocks until it exits or receives a signal
func (p *ProcessManager) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	defer cancel()

	// Create command
	p.cmd = exec.CommandContext(ctx, p.binaryPath, "run", "-c", p.configPath)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr
	p.cmd.Stdin = os.Stdin

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	// Wait for either process exit or signal
	doneChan := make(chan error, 1)
	go func() {
		doneChan <- p.cmd.Wait()
	}()

	select {
	case err := <-doneChan:
		if err != nil {
			return fmt.Errorf("sing-box exited with error: %w", err)
		}
		return nil
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal: %v, shutting down...\n", sig)
		p.Stop()
		<-doneChan // Wait for process to actually exit
		return nil
	case <-ctx.Done():
		p.Stop()
		<-doneChan
		return ctx.Err()
	}
}

// RunWithOutput starts sing-box and captures output
func (p *ProcessManager) RunWithOutput(ctx context.Context, stdout, stderr io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	defer cancel()

	p.cmd = exec.CommandContext(ctx, p.binaryPath, "run", "-c", p.configPath)
	p.cmd.Stdout = stdout
	p.cmd.Stderr = stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	if err := p.cmd.Wait(); err != nil {
		return fmt.Errorf("sing-box exited with error: %w", err)
	}

	return nil
}

// Stop stops the sing-box process gracefully
func (p *ProcessManager) Stop() {
	if p.cmd != nil && p.cmd.Process != nil {
		// Send SIGTERM first for graceful shutdown
		p.cmd.Process.Signal(syscall.SIGTERM)
	}
	if p.cancel != nil {
		p.cancel()
	}
}

// Kill forcefully kills the sing-box process
func (p *ProcessManager) Kill() {
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
}

// IsRunning checks if the process is still running
func (p *ProcessManager) IsRunning() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	// Check if process is still running
	err := p.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// CheckBinary checks if the sing-box binary exists and is executable
func CheckBinary(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("binary not found at %s", path)
		}
		return fmt.Errorf("failed to check binary: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", path)
	}

	// Check if executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("binary at %s is not executable", path)
	}

	return nil
}

// CheckConfig checks if the config file exists
func CheckConfig(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config not found at %s", path)
		}
		return fmt.Errorf("failed to check config: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", path)
	}

	return nil
}

// ValidateConfig validates a sing-box configuration file
func ValidateConfig(binaryPath, configPath string) error {
	cmd := exec.Command(binaryPath, "check", "-c", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("config validation failed: %s", string(output))
	}
	return nil
}
