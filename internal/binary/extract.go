package binary

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nethopper/nethopper/internal/platform"
)

// Extractor handles binary extraction operations
type Extractor struct {
	paths platform.PathProvider
}

// NewExtractor creates a new Extractor with the given path provider
func NewExtractor(paths platform.PathProvider) *Extractor {
	return &Extractor{paths: paths}
}

// ExtractResult contains the result of an extraction operation
type ExtractResult struct {
	Path       string
	Extracted  bool   // true if newly extracted, false if already current
	Hash       string // SHA256 hash of the binary
}

// Extract extracts the embedded binary to the appropriate location
// systemMode: true for server (system paths), false for freenet (user paths)
// Returns the path to the extracted binary
func (e *Extractor) Extract(systemMode bool) (*ExtractResult, error) {
	targetPath := e.paths.BinaryPath(systemMode)
	targetDir := filepath.Dir(targetPath)

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Get embedded binary data
	embeddedData, err := GetEmbeddedBinary()
	if err != nil {
		return nil, err
	}

	embeddedHash := sha256.Sum256(embeddedData)
	hashStr := fmt.Sprintf("%x", embeddedHash)

	// Check if binary exists and is current
	if e.isCurrent(targetPath, embeddedData) {
		return &ExtractResult{
			Path:      targetPath,
			Extracted: false,
			Hash:      hashStr,
		}, nil
	}

	// Write the binary
	perm := os.FileMode(0755)
	if err := os.WriteFile(targetPath, embeddedData, perm); err != nil {
		return nil, fmt.Errorf("failed to write binary to %s: %w", targetPath, err)
	}

	return &ExtractResult{
		Path:      targetPath,
		Extracted: true,
		Hash:      hashStr,
	}, nil
}

// isCurrent checks if the existing binary matches the embedded version
func (e *Extractor) isCurrent(path string, embeddedData []byte) bool {
	existingData, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	return bytes.Equal(existingData, embeddedData)
}

// GetCurrentBinaryPath returns the path where the binary should be for the mode
func (e *Extractor) GetCurrentBinaryPath(systemMode bool) string {
	return e.paths.BinaryPath(systemMode)
}

// IsBinaryInstalled checks if the binary is installed at the expected location
func (e *Extractor) IsBinaryInstalled(systemMode bool) bool {
	path := e.paths.BinaryPath(systemMode)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsBinaryCurrent checks if the installed binary matches the embedded version
func (e *Extractor) IsBinaryCurrent(systemMode bool) (bool, error) {
	path := e.paths.BinaryPath(systemMode)

	embeddedData, err := GetEmbeddedBinary()
	if err != nil {
		return false, err
	}

	return e.isCurrent(path, embeddedData), nil
}

// Remove removes the installed binary
func (e *Extractor) Remove(systemMode bool) error {
	path := e.paths.BinaryPath(systemMode)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary at %s: %w", path, err)
	}

	return nil
}

// GetBinaryVersion attempts to get the version of the installed binary
func (e *Extractor) GetBinaryVersion(systemMode bool) (string, error) {
	path := e.paths.BinaryPath(systemMode)

	if !e.IsBinaryInstalled(systemMode) {
		return "", fmt.Errorf("binary not installed")
	}

	// sing-box version command would be run here
	// For now, return placeholder
	_ = path
	return "unknown", nil
}

// GetTargetPlatform returns a string identifying the current platform
func GetTargetPlatform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
