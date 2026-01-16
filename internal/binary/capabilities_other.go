//go:build !linux

package binary

// SetNetBindCapability is a no-op on non-Linux platforms.
// Server mode (which uses this) is Linux-only and will error at runtime.
func SetNetBindCapability(binaryPath string) error {
	return nil
}
