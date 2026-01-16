//go:build linux && arm64

package binary

import _ "embed"

//go:embed embedded/sing-box-linux-arm64
var embeddedBinary []byte

var embeddedBinaryName = "sing-box-linux-arm64"
