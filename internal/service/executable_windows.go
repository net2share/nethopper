//go:build windows

package service

import "os"

// isExecutable on Windows always returns true since executability is determined
// by file extension (.exe), not by permission bits
func isExecutable(info os.FileInfo) bool {
	return true
}
