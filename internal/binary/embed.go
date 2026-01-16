package binary

import (
	"fmt"
	"runtime"
)

// GetBinaryName returns the expected binary filename for the current platform
func GetBinaryName() string {
	name := fmt.Sprintf("sing-box-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// GetEmbeddedBinary returns the embedded binary data for the current platform
func GetEmbeddedBinary() ([]byte, error) {
	if len(embeddedBinary) == 0 {
		return nil, fmt.Errorf("embedded binary not found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	return embeddedBinary, nil
}

// HasEmbeddedBinary checks if an embedded binary exists for the current platform
func HasEmbeddedBinary() bool {
	return len(embeddedBinary) > 0
}

// GetEmbeddedBinaryName returns the name of the embedded binary
func GetEmbeddedBinaryName() string {
	return embeddedBinaryName
}
