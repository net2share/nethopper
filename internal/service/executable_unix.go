//go:build !windows

package service

import "os"

// isExecutable checks if a file has executable permissions (Unix)
func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0111 != 0
}
