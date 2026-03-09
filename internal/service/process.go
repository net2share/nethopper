package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// ProcessManager manages xray processes for client mode.
type ProcessManager struct {
	binaryPath string
	configPath string
	cmd        *exec.Cmd
	cancel     context.CancelFunc
}

func NewProcessManager(binaryPath, configPath string) *ProcessManager {
	return &ProcessManager{
		binaryPath: binaryPath,
		configPath: configPath,
	}
}

// Run starts xray in foreground and blocks until it exits or receives a signal.
func (p *ProcessManager) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	defer cancel()

	p.cmd = exec.CommandContext(ctx, p.binaryPath, "run", "-c", p.configPath)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr
	p.cmd.Stdin = os.Stdin

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %w", err)
	}

	doneChan := make(chan error, 1)
	go func() {
		doneChan <- p.cmd.Wait()
	}()

	select {
	case err := <-doneChan:
		if err != nil {
			return fmt.Errorf("xray exited with error: %w", err)
		}
		return nil
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal: %v, shutting down...\n", sig)
		p.Stop()
		<-doneChan
		return nil
	case <-ctx.Done():
		p.Stop()
		<-doneChan
		return ctx.Err()
	}
}

func (p *ProcessManager) Stop() {
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Signal(syscall.SIGTERM)
	}
	if p.cancel != nil {
		p.cancel()
	}
}
