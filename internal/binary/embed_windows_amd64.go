//go:build windows && amd64

package binary

import _ "embed"

//go:embed embedded/sing-box-windows-amd64.exe
var embeddedBinary []byte

var embeddedBinaryName = "sing-box-windows-amd64.exe"
