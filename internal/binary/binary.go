package binary

import (
	"runtime"

	"github.com/net2share/go-corelib/binman"
)

// XrayDef defines the Xray binary for binman.
var XrayDef = binman.BinaryDef{
	Name:          xrayBinaryName(),
	EnvOverride:   "NETHOPPER_XRAY_PATH",
	URLPattern:    "https://github.com/XTLS/Xray-core/releases/download/{version}/Xray-{xrayplatform}.zip",
	PinnedVersion: "v25.3.6",
	ArchiveType:   "zip",
	ChecksumURL:   "https://github.com/XTLS/Xray-core/releases/download/{version}/Xray-{xrayplatform}.zip.dgst",
	Platforms: map[string][]string{
		"linux":   {"amd64", "arm64"},
		"darwin":  {"amd64", "arm64"},
		"windows": {"amd64"},
	},
	ArchMappings: map[string]binman.ArchMapping{
		"xrayplatform": {
			"linux/amd64":   "linux-64",
			"linux/arm64":   "linux-arm64-v8a",
			"darwin/amd64":  "macos-64",
			"darwin/arm64":  "macos-arm64-v8a",
			"windows/amd64": "windows-64",
		},
	},
}

func xrayBinaryName() string {
	if runtime.GOOS == "windows" {
		return "xray.exe"
	}
	return "xray"
}

// NewServerManager creates a binman manager for server (root) installation.
func NewServerManager() *binman.Manager {
	return binman.NewManager("/usr/local/bin",
		binman.WithSystemPaths([]string{"/usr/local/bin", "/usr/bin"}),
	)
}

// NewClientManager creates a binman manager for client (user) installation.
func NewClientManager(binDir string) *binman.Manager {
	return binman.NewManager(binDir)
}
